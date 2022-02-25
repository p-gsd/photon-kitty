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
	childWidth       int
	childHeight      int
}

func (g *Grid) Draw(ctx Context, s tcell.Screen, sixelBuf *bytes.Buffer, full bool) (statusBarText richtext) {
	margin := (int(ctx.Width) % g.Columns) / 2
	if photon.SelectedCard == nil && photon.VisibleCards != nil {
		photon.SelectedCardPos = image.Point{
			X: g.FirstChildIndex % g.Columns,
			Y: g.FirstChildIndex / g.Columns,
		}
		if g.FirstChildIndex >= len(photon.VisibleCards) {
			g.FirstChildIndex = len(photon.VisibleCards) - 1
		}
		photon.SelectedCard = photon.VisibleCards[g.FirstChildIndex]
	}
	for i := g.FirstChildIndex; i < len(photon.VisibleCards); i++ {
		child := photon.VisibleCards[i]
		chctx := Context{
			WinSize:     ctx.WinSize,
			X:           margin + (i%g.Columns)*g.childWidth,
			Y:           g.FirstChildOffset + ((i-g.FirstChildIndex)/g.Columns)*g.childHeight,
			Width:       g.childWidth,
			Height:      g.childHeight,
			XCellPixels: ctx.XCellPixels,
			YCellPixels: ctx.YCellPixels,
		}
		if chctx.Y >= int(ctx.Height) {
			break
		}
		g.LastChildIndex = i
		g.RowsCount = i / g.Columns
		g.LastChildOffset = chctx.Y + g.childHeight - int(ctx.Height)
		getCard(child).Draw(chctx, s, sixelBuf, full)
	}
	//clear remaining card space
	if g.LastChildIndex == len(photon.VisibleCards)-1 {
		for i := g.LastChildIndex + 1; i < (g.RowsCount+1)*g.Columns; i++ {
			X := margin + (i%g.Columns)*g.childWidth
			Y := g.FirstChildOffset + ((i-g.FirstChildIndex)/g.Columns)*g.childHeight
			fillArea(s, image.Rect(X, Y, X+g.childWidth, Y+g.childHeight), ' ')
		}
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
			WinSize:     ctx.WinSize,
			Width:       g.childWidth,
			Height:      g.childHeight,
			XCellPixels: ctx.XCellPixels,
			YCellPixels: ctx.YCellPixels,
		}
		getCard(child).DownloadImage(chctx, s)
	}
	//status bar text - scroll percentage
	above := (g.FirstChildIndex/g.Columns)*g.childHeight - g.FirstChildOffset
	allRows := int(math.Ceil(float64(len(photon.VisibleCards)) / float64(g.Columns)))
	below := (allRows-(g.LastChildIndex/g.Columns)-1)*g.childHeight + g.LastChildOffset
	statusBarText = richtext{{Text: scrollPercentage(above, below), Style: tcell.StyleDefault}}
	return
}

func (g *Grid) Resize(ctx Context) {
	g.childWidth = ctx.Width / g.Columns
	g.childHeight = g.childWidth * ctx.XCellPixels / ctx.YCellPixels
}

func (g *Grid) ClearImages() {
	for _, ch := range photon.Cards {
		getCard(ch).ClearImage()
	}
}

func (g *Grid) ClearCardsPosition() {
	for _, card := range photon.Cards {
		getCard(card).previousImagePos = image.Point{-2, -2}
	}
}

func (g *Grid) Scroll(d int) {
	defer g.selectedChildRefresh()
	cardDiff := (d / g.childHeight) * g.Columns
	cellDiff := d % g.childHeight
	allRows := math.Ceil(float64(len(photon.VisibleCards)) / float64(g.Columns))
	if math.Ceil(float64(g.LastChildIndex+cardDiff)/float64(g.Columns)) >= allRows {
		g.FirstChildIndex = len(photon.VisibleCards) - (g.LastChildIndex - g.FirstChildIndex) - 1
		g.LastChildIndex = len(photon.VisibleCards) - 1
		return
	}
	g.FirstChildIndex += cardDiff
	g.FirstChildOffset += -cellDiff
	g.LastChildIndex += cardDiff
	g.LastChildOffset += -cellDiff
	for -g.FirstChildOffset > g.childHeight {
		g.FirstChildIndex += g.Columns
		g.LastChildIndex += g.Columns
		g.FirstChildOffset += g.childHeight
		g.LastChildOffset += g.childHeight
	}
	for g.FirstChildOffset > 0 {
		g.FirstChildIndex -= g.Columns
		g.LastChildIndex -= g.Columns
		g.FirstChildOffset -= g.childHeight
		g.LastChildOffset -= g.childHeight
	}
	if g.FirstChildIndex < 0 {
		g.FirstChildIndex = 0
		g.FirstChildOffset = 0
	}
	if g.FirstChildIndex/g.Columns > photon.SelectedCardPos.Y {
		photon.SelectedCardPos.Y = g.FirstChildIndex / g.Columns
		if g.FirstChildOffset < 0 {
			photon.SelectedCardPos.Y++
		}
	}
	if g.LastChildIndex/g.Columns < photon.SelectedCardPos.Y {
		photon.SelectedCardPos.Y = g.LastChildIndex / g.Columns
		if g.LastChildOffset > 0 {
			photon.SelectedCardPos.Y--
		}
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
	if len(photon.VisibleCards)-1 < index {
		photon.SelectedCard = nil
		return
	}
	photon.SelectedCard = photon.VisibleCards[index]
}
