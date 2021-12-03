package main

import (
	"git.sr.ht/~ghost08/libphoton/states"
	"github.com/gdamore/tcell/v2"
)

func drawCommand(ctx Context, s tcell.Screen) {
	for i := 0; i < int(ctx.Cols); i++ {
		s.SetContent(i, int(ctx.Rows)-1, ' ', nil, tcell.StyleDefault)
	}
	drawLines(
		s,
		0,
		int(ctx.Rows-1),
		int(ctx.Cols),
		1,
		command,
		tcell.StyleDefault,
	)
	if commandFocus {
		s.ShowCursor(len(command), int(ctx.Rows-1))
	} else {
		s.HideCursor()
	}
}

func commandInput(s tcell.Screen, ev *tcell.EventKey) bool {
	if cb.State() != states.Search || !commandFocus {
		return false
	}
	if ev.Key() != tcell.KeyRune {
		return false
	}
	if ev.Key() == tcell.KeyBackspace {
		command = command[:len(command)-1]
	} else {
		command += string(ev.Rune())
	}
	photon.SearchQuery(command[1:])
	photon.SelectedCard = nil
	if len(photon.VisibleCards) > 0 {
		photon.SelectedCard = photon.VisibleCards[0]
	}
	s.Clear()
	redraw(true)
	return true
}
