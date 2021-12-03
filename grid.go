package main

import (
	"bytes"
	"image"
	"math"

	"github.com/gdamore/tcell/v2"
)

type Grid struct {
	Columns          int
	FirstChildIndex  int
	FirstChildOffset int
	LastChildIndex   int
	LastChildOffset  int
	RowsCount        int //count of visible rows on the screen
}

func (g *Grid) Draw(ctx Context, s tcell.Screen) (buf *bytes.Buffer) {
	buf = bytes.NewBuffer(nil)
	w, h := s.Size()
	margin := (w % g.Columns) / 2
	childWidth := w / g.Columns
	childHeight := int(float32(childWidth) / 2.2)
	if photon.SelectedCard == nil && photon.VisibleCards != nil {
		photon.SelectedCardPos = image.Point{
			X: g.FirstChildIndex % g.Columns,
			Y: g.FirstChildIndex / g.Columns,
		}
		photon.SelectedCard = photon.VisibleCards[g.FirstChildIndex]
		getCard(photon.SelectedCard).Select()
	}
	for i := g.FirstChildIndex; i < len(photon.VisibleCards); i++ {
		child := photon.VisibleCards[i]
		chctx := Context{
			WinSize: ctx.WinSize,
			X:       margin + (i%g.Columns)*childWidth,
			Y:       g.FirstChildOffset + ((i-g.FirstChildIndex)/g.Columns)*childHeight,
			Width:   childWidth,
			Height:  childHeight,
		}
		if chctx.Y >= h {
			break
		}
		g.LastChildIndex = i
		g.RowsCount = i / g.Columns
		g.LastChildOffset = chctx.Y + childHeight - h
		getCard(child).Draw(chctx, s, buf)
	}
	//set all not visible cards previous position outside
	for i := 0; i < len(photon.VisibleCards); i++ {
		if i >= g.FirstChildIndex && i <= g.LastChildIndex {
			continue
		}
		getCard(photon.VisibleCards[i]).previousImagePos = image.Point{-2, -2}
	}
	//download next screen of images
	for i := g.LastChildIndex + 1; i < len(photon.VisibleCards) && i < g.LastChildIndex+(g.RowsCount*g.Columns)+1; i++ {
		child := photon.VisibleCards[i]
		chctx := Context{
			WinSize: ctx.WinSize,
			Width:   childWidth,
			Height:  childHeight,
		}
		getCard(child).DownloadImage(chctx, s)
	}
	return buf
}

func (g *Grid) ClearImages() {
	for _, ch := range photon.Cards {
		getCard(ch).ClearImage()
	}
}

func (g *Grid) SelectedChildMoveLeft() {
	if photon.SelectedCard == nil {
		return
	}
	defer g.selectedChildRefresh()
	if photon.SelectedCardPos.X == 0 {
		if photon.SelectedCardPos.Y == 0 {
			return
		}
		photon.SelectedCardPos.Y--
		if g.FirstChildIndex/g.Columns == photon.SelectedCardPos.Y {
			g.FirstChildOffset = 0
		}
		if g.FirstChildIndex/g.Columns > photon.SelectedCardPos.Y {
			g.FirstChildIndex = photon.SelectedCardPos.Y * g.Columns
		}
		photon.SelectedCardPos.X = g.Columns - 1
		return
	}
	photon.SelectedCardPos.X--
}

func (g *Grid) SelectedChildMoveRight() {
	if photon.SelectedCard == nil {
		return
	}
	defer g.selectedChildRefresh()
	if photon.SelectedCardPos.Y == int(math.Ceil(float64(len(photon.VisibleCards))/float64(g.Columns)))-1 {
		if photon.SelectedCardPos.X == g.Columns-1 || len(photon.VisibleCards) == (photon.SelectedCardPos.Y*g.Columns+photon.SelectedCardPos.X+1) {
			return
		}
		photon.SelectedCardPos.X++
		return
	}
	if photon.SelectedCardPos.X == g.Columns-1 {
		photon.SelectedCardPos.Y++
		photon.SelectedCardPos.X = 0
		if g.LastChildIndex/g.Columns < photon.SelectedCardPos.Y {
			g.FirstChildIndex += g.Columns
		}
		if g.LastChildIndex/g.Columns == photon.SelectedCardPos.Y {
			g.FirstChildOffset -= g.LastChildOffset
			g.LastChildOffset = 0
		}
		return
	}
	photon.SelectedCardPos.X++
}

func (g *Grid) SelectedChildMoveDown() {
	if photon.SelectedCard == nil {
		return
	}
	defer g.selectedChildRefresh()
	maxRow := int(math.Ceil(float64(len(photon.VisibleCards))/float64(g.Columns))) - 1
	if photon.SelectedCardPos.Y == maxRow {
		return
	}
	photon.SelectedCardPos.Y++
	if photon.SelectedCardPos.Y == maxRow {
		photon.SelectedCardPos.X = min(photon.SelectedCardPos.X, len(photon.VisibleCards)%g.Columns-1)
	}
	if g.LastChildIndex/g.Columns == photon.SelectedCardPos.Y {
		g.FirstChildOffset -= g.LastChildOffset
		g.LastChildOffset = 0
	}
	if g.LastChildIndex/g.Columns < photon.SelectedCardPos.Y {
		g.FirstChildIndex += g.Columns
	}
}

func (g *Grid) SelectedChildMoveUp() {
	if photon.SelectedCard == nil {
		return
	}
	defer g.selectedChildRefresh()
	if photon.SelectedCardPos.Y == 0 {
		return
	}
	photon.SelectedCardPos.Y--
	if g.FirstChildIndex/g.Columns == photon.SelectedCardPos.Y {
		g.FirstChildOffset = 0
	}
	if g.FirstChildIndex/g.Columns > photon.SelectedCardPos.Y {
		g.FirstChildIndex = photon.SelectedCardPos.Y * g.Columns
	}
}

func (g *Grid) selectedChildRefresh() {
	if photon.SelectedCardPos.Y < 0 {
		photon.SelectedCardPos.Y = 0
	}
	if photon.SelectedCardPos.X < 0 {
		photon.SelectedCardPos.X = 0
	}
	index := photon.SelectedCardPos.Y*g.Columns + photon.SelectedCardPos.X
	c := getCard(photon.SelectedCard)
	if len(photon.VisibleCards)-1 < index {
		c.Unselect()
		photon.SelectedCard = nil
		return
	}
	c.Unselect()
	photon.SelectedCard = photon.VisibleCards[index]
	getCard(photon.SelectedCard).Select()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
