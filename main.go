package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"git.sr.ht/~ghost08/libphoton"
	"git.sr.ht/~ghost08/libphoton/keybindings"
	"git.sr.ht/~ghost08/libphoton/states"

	"github.com/alecthomas/kong"
	"github.com/gdamore/tcell/v2"
	"golang.design/x/clipboard"
)

var CLI struct {
	Extractor    string       `optional:"" default:"youtube-dl --get-url %" help:"command for media link extraction (item link is substituted for %)" env:"PHOTON_EXTRACTOR"`
	VideoCmd     string       `optional:"" default:"mpv $" help:"set default command for opening the item media link in a video player (media link is substituted for %, direct item link is substituted for $, if no % or $ is provided, photon will download the data and pipe it to the stdin of the command)" env:"PHOTON_VIDEOCMD"`
	ImageCmd     string       `optional:"" default:"imv -" help:"set default command for opening the item media link in a image viewer (media link is substituted for %, direct item link is substituted for $, if no % or $ is provided, photon will download the data and pipe it to the stdin of the command)" env:"PHOTON_IMAGECMD"`
	TorrentCmd   string       `optional:"" default:"mpv %" help:"set default command for opening the item media link in a torrent downloader (media link is substituted for %, if link is a torrent file, photon will download it, and substitute the torrent file path for %)" env:"PHOTON_TORRENTCMD"`
	HTTPSettings HTTPSettings `embed:""`
	Paths        []string     `arg:"" optional:"" help:"RSS/Atom urls, config path, or - for stdin"`
	DownloadPath string       `optional:"" default:"$HOME/Downloads" help:"the default download path"`
}

var (
	photon       *libphoton.Photon
	cb           Callbacks
	command      string
	commandFocus bool
)

func main() {
	f, _ := os.Create("/tmp/photon.log")
	log.SetOutput(f)
	defer f.Close()
	//args
	kong.Parse(&CLI,
		kong.Name("photon"),
		kong.Description("Fast RSS reader as light as a photon"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	if len(CLI.Paths) == 0 {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		defaultConf := filepath.Join(usr.HomeDir, ".config", "photon", "config")
		if _, err := os.Stat(defaultConf); os.IsNotExist(err) {
			log.Fatal(err)
		}
		CLI.Paths = []string{defaultConf}
	}

	//photon
	grid := &Grid{Columns: 5}
	cb = Callbacks{grid: grid}
	var err error
	photon, err = libphoton.New(
		cb,
		CLI.Paths,
		libphoton.WithHTTPClient(CLI.HTTPSettings.Client()),
		libphoton.WithMediaExtractor(CLI.Extractor),
		libphoton.WithMediaVideoCmd(CLI.VideoCmd),
		libphoton.WithMediaImageCmd(CLI.ImageCmd),
		libphoton.WithMediaTorrentCmd(CLI.TorrentCmd),
		libphoton.WithDownloadPath(CLI.DownloadPath),
	)
	if err != nil {
		log.Fatal(err)
	}

	//tui
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

	go func() {
		photon.RefreshFeed()
		redraw(true)
	}()

	defaultKeyBindings(s, grid, &quit)

	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if commandInput(s, ev) {
					continue
				}
				ke := newKeyEvent(ev)
				photon.KeyBindings.Run(ke)
			case *tcell.EventResize:
				s.Clear()
				grid.ClearImages()
				imageProcClear()
				ctx, quit = WithCancel(Background())
				redraw(true)
			}
		}
	}()

	ctx.Width -= 1
	var fullRedraw bool
	var sixelBuf *bytes.Buffer
	for {
		switch cb.State() {
		case states.Normal, states.Search:
			sixelBuf = grid.Draw(ctx, s)
			drawCommand(ctx, s)
		case states.Article:
			sixelBuf = openedArticle.Draw(ctx, s)
		}
		if fullRedraw {
			s.Sync()
		} else {
			s.Show()
		}
		if sixelBuf != nil && sixelBuf.Len() > 0 {
			os.Stdout.Write(sixelBuf.Bytes())
		}
		select {
		case <-ctx.Done():
			return
		case fullRedraw = <-redrawCh:
		}
	}
}

var redrawCh = make(chan bool, 1024)

func redraw(full bool) {
	redrawCh <- full
}

func redrawWorker() {
	redrawReq := make(chan bool)
	go func() {
		var timestamp time.Time
		for f := range redrawCh {
			if time.Now().Sub(timestamp) < time.Second/30 {
				continue
			}
			redrawReq <- f
		}
	}()
}

func newKeyEvent(e *tcell.EventKey) keybindings.KeyEvent {
	var mod keybindings.Modifiers
	switch {
	case e.Modifiers()&tcell.ModCtrl != 0:
		mod = keybindings.ModCtrl
	case e.Modifiers()&tcell.ModShift != 0:
		mod = keybindings.ModShift
	case e.Modifiers()&tcell.ModAlt != 0:
		mod = keybindings.ModAlt
	case e.Modifiers()&tcell.ModMeta != 0:
		mod = keybindings.ModSuper
	}

	var r rune
	switch {
	case e.Key() == tcell.KeyBackspace:
		return keybindings.KeyEvent{Key: '\u0008'}
	case e.Key() == tcell.KeyTab:
		return keybindings.KeyEvent{Key: '\t'}
	case e.Key() == tcell.KeyEsc:
		return keybindings.KeyEvent{Key: '\u00b1'}
	case e.Key() == tcell.KeyEnter:
		return keybindings.KeyEvent{Key: '\n'}
	case e.Key() == tcell.KeyRune:
		if unicode.IsUpper(e.Rune()) {
			mod = keybindings.ModShift
		}
		r = unicode.ToLower(e.Rune())
		return keybindings.KeyEvent{Key: r, Modifiers: mod}
	default:
		s, ok := tcell.KeyNames[e.Key()]
		if ok && strings.HasPrefix(s, "Ctrl-") {
			s = s[5:]
			r, _ = utf8.DecodeLastRuneInString(s)
			r = unicode.ToLower(r)
		}
		return keybindings.KeyEvent{Key: r, Modifiers: mod}
	}
}

func defaultKeyBindings(s tcell.Screen, grid *Grid, quit *context.CancelFunc) {
	//NormalState
	photon.KeyBindings.Add(states.Normal, "q", func() error {
		if quit != nil {
			q := *quit
			quit = nil
			q()
		}
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "<enter>", func() error {
		photon.SelectedCard.OpenArticle()
		for i := 0; i < len(photon.VisibleCards); i++ {
			getCard(photon.VisibleCards[i]).previousImagePos = image.Point{-2, -2}
		}
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "p", func() error {
		photon.SelectedCard.RunMedia()
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "o", func() error {
		photon.SelectedCard.OpenBrowser()
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "<esc>", func() error {
		if command == "" {
			return nil
		}
		command = ""
		commandFocus = false
		photon.VisibleCards = photon.Cards
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "=", func() error {
		grid.Columns++
		grid.ClearImages()
		imageProcClear()
		redraw(true)
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "-", func() error {
		if grid.Columns == 1 {
			return nil
		}
		grid.Columns--
		grid.ClearImages()
		imageProcClear()
		redraw(true)
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "/", func() error {
		if command != "" && !commandFocus {
			commandFocus = true
			redraw(false)
			return nil
		}
		command = "/"
		commandFocus = true
		redraw(false)
		return nil
	})
	//copy item link
	photon.KeyBindings.Add(states.Normal, "yy", func() error {
		if photon.SelectedCard == nil {
			return nil
		}
		clipboard.Write(clipboard.FmtText, []byte(photon.SelectedCard.Item.Link))
		return nil
	})
	//copy item image
	photon.KeyBindings.Add(states.Normal, "yi", func() error {
		if photon.SelectedCard == nil {
			return nil
		}
		if photon.SelectedCard.ItemImage == nil {
			return nil
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, photon.SelectedCard.ItemImage); err != nil {
			return fmt.Errorf("encoding image: %w", err)
		}
		clipboard.Write(clipboard.FmtImage, buf.Bytes())
		return nil
	})
	//download media
	photon.KeyBindings.Add(states.Normal, "dm", func() error {
		photon.SelectedCard.DownloadMedia()
		return nil
	})
	//download link
	photon.KeyBindings.Add(states.Normal, "dl", func() error {
		photon.SelectedCard.DownloadLink()
		return nil
	})
	//download image
	photon.KeyBindings.Add(states.Normal, "di", func() error {
		photon.SelectedCard.DownloadImage()
		return nil
	})
	//move selectedCard
	photon.KeyBindings.Add(states.Normal, "h", func() error {
		cb.SelectedCardMoveLeft()
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "l", func() error {
		cb.SelectedCardMoveRight()
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "j", func() error {
		cb.SelectedCardMoveDown()
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "k", func() error {
		cb.SelectedCardMoveUp()
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "<ctrl>d", func() error {
		/*TODO
		list.Position.First += int(float32(list.Position.Count) / 2)
		redraw()
		*/
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "<ctrl>u", func() error {
		/*TODO
		list.Position.First -= int(float32(list.Position.Count) / 2)
		if list.Position.First < 0 {
			list.Position.First = 0
		}
		redraw()
		*/
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "<ctrl>f", func() error {
		/*TODO
		list.Position.First += list.Position.Count
		redraw()
		*/
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "<ctrl>b", func() error {
		/*TODO
		list.Position.First -= list.Position.Count
		if list.Position.First < 0 {
			list.Position.First = 0
		}
		redraw()
		*/
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "gg", func() error {
		grid.FirstChildIndex = 0
		grid.FirstChildOffset = 0
		photon.SelectedCardPos.Y = 0
		redraw(true)
		return nil
	})
	photon.KeyBindings.Add(states.Normal, "<shift>g", func() error {
		log.Println("<shift>g", grid.RowsCount)
		grid.FirstChildIndex = len(photon.VisibleCards) - grid.RowsCount
		photon.SelectedCardPos.Y = len(photon.VisibleCards)/grid.Columns - 1
		redraw(true)
		return nil
	})

	//SearchState
	photon.KeyBindings.Add(states.Search, "<enter>", func() error {
		commandFocus = false
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Search, "<esc>", func() error {
		command = ""
		commandFocus = false
		photon.VisibleCards = photon.Cards
		redraw(false)
		return nil
	})

	//ArticleState
	photon.KeyBindings.Add(states.Article, "<esc>", func() error {
		openedArticle = nil
		photon.OpenedArticle = nil
		s.Clear()
		redraw(true)
		return nil
	})
	photon.KeyBindings.Add(states.Article, "q", func() error {
		openedArticle = nil
		photon.OpenedArticle = nil
		s.Clear()
		redraw(true)
		return nil
	})
	photon.KeyBindings.Add(states.Article, "j", func() error {
		if openedArticle == nil {
			return nil
		}
		openedArticle.scroll(1)
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Article, "k", func() error {
		if openedArticle == nil {
			return nil
		}
		openedArticle.scroll(-1)
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Article, "gg", func() error {
		if openedArticle == nil {
			return nil
		}
		openedArticle.firstLine = 0
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Article, "<shift>g", func() error {
		if openedArticle == nil {
			return nil
		}
		openedArticle.firstLine = openedArticle.lastLine - len(openedArticle.contentLines)
		redraw(false)
		return nil
	})
	photon.KeyBindings.Add(states.Article, "<ctrl>d", func() error {
		if openedArticle == nil {
			return nil
		}
		openedArticle.scroll((openedArticle.lastLine - openedArticle.firstLine) / 2)
		redraw(true)
		return nil
	})
	photon.KeyBindings.Add(states.Article, "<ctrl>u", func() error {
		if openedArticle == nil {
			return nil
		}
		openedArticle.scroll(-(openedArticle.lastLine - openedArticle.firstLine) / 2)
		redraw(true)
		return nil
	})
}
