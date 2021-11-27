package main

import (
	"bytes"
	"image"
	"io"
	"math"

	"github.com/gdamore/tcell/v2"
)

type Grid struct {
	Columns  int
	Children []Child

	selectedChildPos image.Point
	selectedChild    Child

	FirstChildIndex  int
	FirstChildOffset int
	LastChildIndex   int
	LastChildOffset  int
}

type Selectable interface {
	Selected() bool
	Select()
	Unselect()
}

type Child interface {
	Selectable
	Draw(ctx Context, s tcell.Screen, w io.Writer)
	ClearImage()
}

func (g *Grid) Draw(ctx Context, s tcell.Screen) *bytes.Buffer {
	w, h := s.Size()
	margin := (w % g.Columns) / 2
	childWidth := w / g.Columns
	childHeight := childWidth / 2
	if g.selectedChild == nil && g.Children != nil {
		g.selectedChildPos = image.Point{
			X: g.FirstChildIndex % g.Columns,
			Y: g.FirstChildIndex / g.Columns,
		}
		g.selectedChild = g.Children[g.FirstChildIndex]
		g.selectedChild.Select()
	}
	var buf bytes.Buffer
	for i := g.FirstChildIndex; i < len(g.Children); i++ {
		child := g.Children[i]
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
		g.LastChildOffset = chctx.Y + childHeight - h
		child.Draw(chctx, s, &buf)
	}
	return &buf
}

func (g *Grid) ClearImages() {
	for _, ch := range g.Children {
		ch.ClearImage()
	}
}

func (g *Grid) SelectedChildMoveLeft() {
	if g.selectedChild == nil {
		return
	}
	defer g.selectedChildRefresh()
	if g.selectedChildPos.X == 0 {
		if g.selectedChildPos.Y == 0 {
			return
		}
		g.selectedChildPos.Y--
		if g.FirstChildIndex/g.Columns == g.selectedChildPos.Y {
			g.FirstChildOffset = 0
		}
		if g.FirstChildIndex/g.Columns > g.selectedChildPos.Y {
			g.FirstChildIndex = g.selectedChildPos.Y * g.Columns
		}
		g.selectedChildPos.X = g.Columns - 1
		return
	}
	g.selectedChildPos.X--
}

func (g *Grid) SelectedChildMoveRight() {
	if g.selectedChild == nil {
		return
	}
	defer g.selectedChildRefresh()
	if g.selectedChildPos.Y == int(math.Ceil(float64(len(g.Children))/float64(g.Columns)))-1 {
		if g.selectedChildPos.X == g.Columns-1 || len(g.Children) == (g.selectedChildPos.Y*g.Columns+g.selectedChildPos.X+1) {
			return
		}
		g.selectedChildPos.X++
		return
	}
	if g.selectedChildPos.X == g.Columns-1 {
		g.selectedChildPos.Y++
		g.selectedChildPos.X = 0
		if g.LastChildIndex/g.Columns < g.selectedChildPos.Y {
			g.FirstChildIndex += g.Columns
		}
		if g.LastChildIndex/g.Columns == g.selectedChildPos.Y {
			g.FirstChildOffset -= g.LastChildOffset
			g.LastChildOffset = 0
		}
		return
	}
	g.selectedChildPos.X++
}

func (g *Grid) SelectedChildMoveDown() {
	if g.selectedChild == nil {
		return
	}
	defer g.selectedChildRefresh()
	maxRow := int(math.Ceil(float64(len(g.Children))/float64(g.Columns))) - 1
	if g.selectedChildPos.Y == maxRow {
		return
	}
	g.selectedChildPos.Y++
	if g.selectedChildPos.Y == maxRow {
		g.selectedChildPos.X = min(g.selectedChildPos.X, len(g.Children)%g.Columns-1)
	}
	if g.LastChildIndex/g.Columns == g.selectedChildPos.Y {
		g.FirstChildOffset -= g.LastChildOffset
		g.LastChildOffset = 0
	}
	if g.LastChildIndex/g.Columns < g.selectedChildPos.Y {
		g.FirstChildIndex += g.Columns
	}
}

func (g *Grid) SelectedChildMoveUp() {
	if g.selectedChild == nil {
		return
	}
	defer g.selectedChildRefresh()
	if g.selectedChildPos.Y == 0 {
		return
	}
	g.selectedChildPos.Y--
	if g.FirstChildIndex/g.Columns == g.selectedChildPos.Y {
		g.FirstChildOffset = 0
	}
	if g.FirstChildIndex/g.Columns > g.selectedChildPos.Y {
		g.FirstChildIndex = g.selectedChildPos.Y * g.Columns
	}
}

func (g *Grid) selectedChildRefresh() {
	if g.selectedChildPos.Y < 0 {
		g.selectedChildPos.Y = 0
	}
	if g.selectedChildPos.X < 0 {
		g.selectedChildPos.X = 0
	}
	index := g.selectedChildPos.Y*g.Columns + g.selectedChildPos.X
	if len(g.Children)-1 < index {
		g.selectedChild.Unselect()
		g.selectedChild = nil
		return
	}
	g.selectedChild.Unselect()
	g.selectedChild = g.Children[index]
	g.selectedChild.Select()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (g *Grid) EventKey(s tcell.Screen, ev *tcell.EventKey) {
	switch ev.Rune() {
	case '-':
		if g.Columns == 1 {
			return
		}
		g.Columns -= 1
		g.ClearImages()
		redraw()
	case '=':
		g.Columns += 1
		g.ClearImages()
		redraw()
	case 'h':
		g.SelectedChildMoveLeft()
		redraw()
	case 'l':
		g.SelectedChildMoveRight()
		redraw()
	case 'j':
		g.SelectedChildMoveDown()
		redraw()
	case 'k':
		g.SelectedChildMoveUp()
		redraw()
	}
}
