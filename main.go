package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)

var (
	redrawCh = make(chan struct{})
)

func redraw() {
	redrawCh <- struct{}{}
}

func main() {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err = s.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer s.Fini()

	ctx, quit := WithCancel(Background())
	grid := &Grid{Columns: 5}

	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape:
					quit()
					return
				case tcell.KeyCtrlL:
					s.Sync()
				}
				switch ev.Rune() {
				case 'q':
					quit()
					return
				}
				grid.EventKey(s, ev)
			case *tcell.EventResize:
				s.Sync()
				grid.ClearImages()
				ctx, quit = WithCancel(Background())
				redraw()
			}
		}
	}()

	addTestDataToGrid(grid)

	for {
		s.Clear()
		sixelBuf := grid.Draw(ctx, s)
		s.Sync()
		//io.Copy(os.Stdout, sixelBuf)
		os.Stdout.Write(sixelBuf.Bytes())
		select {
		case <-ctx.Done():
			return
		case <-redrawCh:
		}
	}
}
