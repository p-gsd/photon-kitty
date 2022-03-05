package main

import (
	"image"

	"git.sr.ht/~ghost08/photon/lib/states"
	"github.com/gdamore/tcell/v2"
)

func drawCommand(ctx Context, s tcell.Screen) {
	for i := 0; i < int(ctx.Cols); i++ {
		s.SetContent(i, int(ctx.Rows)-1, ' ', nil, tcell.StyleDefault)
	}
	drawLine(
		s,
		0,
		int(ctx.Rows-1),
		int(ctx.Cols),
		command,
		tcell.StyleDefault,
	)
}

func commandInput(s tcell.Screen, ev *tcell.EventKey) bool {
	if cb.State() != states.Search || !commandFocus {
		return false
	}
	if ev.Key() != tcell.KeyRune && ev.Key() != tcell.KeyBackspace && ev.Key() != tcell.KeyBackspace2 {
		return false
	}
	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		command = command[:len(command)-1]
	} else {
		command += string(ev.Rune())
	}
	if len(command) > 0 {
		photon.SearchQuery(command[1:])
		SelectedCard = nil
		if len(photon.VisibleCards) > 0 {
			SelectedCard = photon.VisibleCards[0]
			SelectedCardPos = image.Point{}
		}
	} else {
		commandFocus = false
		photon.VisibleCards = photon.Cards
		SelectedCard = photon.VisibleCards[0]
		SelectedCardPos = image.Point{}
	}
	s.Clear()
	redraw(true)
	return true
}
