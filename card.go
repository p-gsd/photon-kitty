package main

import (
	"image"
	"io"
	"time"

	"git.sr.ht/~ghost08/photon/lib"
	"github.com/gdamore/tcell/v2"
	htime "github.com/sbani/go-humanizer/time"
)

var cards = make(map[*lib.Card]*Card)

func getCard(card *lib.Card) *Card {
	c, ok := cards[card]
	if !ok {
		c = &Card{
			Card: card,
		}
		cards[card] = c
	}
	return c
}

const (
	headerHeight  = 4
	selectedColor = tcell.ColorGray
)

type Card struct {
	*lib.Card
	sixelData         *Sixel
	scaledImageBounds image.Rectangle
	//isOnScreen        func(*lib.Card)
	previousImagePos image.Point
	previousSelected bool
}

func (c *Card) Draw(ctx Context, s tcell.Screen, w io.Writer, full bool) {
	imageWidthInCells := c.scaledImageBounds.Dx() / ctx.XCellPixels()
	offset := (ctx.Width - imageWidthInCells) / 2
	newImagePos := image.Point{ctx.X + 1 + offset, ctx.Y + 1}
	selected := c.Card == photon.SelectedCard
	if !full && c.previousImagePos.Eq(newImagePos) && selected == c.previousSelected {
		return
	}
	style := tcell.StyleDefault
	if c.Foreground != -1 {
		style = style.Foreground(tcell.ColorValid + tcell.Color(c.Foreground))
	}
	if c.Background != -1 {
		style = style.Background(tcell.ColorValid + tcell.Color(c.Background))
	}
	if selected {
		style = tcell.StyleDefault.Background(selectedColor)
	}
	if c.Item.Image == nil {
		for x := ctx.X; x < ctx.Width+ctx.X; x++ {
			for y := ctx.Y; y < ctx.Height+ctx.Y; y++ {
				s.SetContent(x, y, ' ', nil, style)
			}
		}
		drawLinesWordwrap(s, ctx.X+1, ctx.Y, ctx.Width-3, 2, c.Item.Title, style.Bold(true))
		drawLine(s, ctx.X+1, ctx.Y+2, ctx.Width-3, c.Feed.Title, style.Italic(true))
		drawLine(s, ctx.X+1, ctx.Y+3, ctx.Width-3, htime.Difference(time.Now(), *c.Item.PublishedParsed), style.Italic(true))
		drawLinesWordwrap(s, ctx.X+1, ctx.Y+headerHeight+1, ctx.Width-3, ctx.Height-headerHeight-2, c.Item.Custom["simpleContent"], style)
		return
	}

	//header
	for x := ctx.X; x < ctx.Width+ctx.X; x++ {
		for y := ctx.Height - headerHeight + ctx.Y; y < ctx.Height+ctx.Y; y++ {
			s.SetContent(x, y, ' ', nil, style)
		}
	}
	drawLinesWordwrap(s, ctx.X+1, ctx.Height-headerHeight+ctx.Y, ctx.Width-3, 2, c.Item.Title, style.Bold(true))
	drawLine(s, ctx.X+1, ctx.Height-headerHeight+ctx.Y+2, ctx.Width-3, c.Feed.Title, style.Italic(true))
	drawLine(s, ctx.X+1, ctx.Height-headerHeight+ctx.Y+3, ctx.Width-3, htime.Difference(time.Now(), *c.Item.PublishedParsed), style.Italic(true))

	if c.DownloadImage(ctx, s) {
		c.previousImagePos = image.Point{-2, -2}
		c.swapImageRegion(ctx, s)
		return
	}
	if c.sixelData == nil {
		c.previousImagePos = image.Point{-2, -2}
		c.swapImageRegion(ctx, s)
		return
	}
	if !c.previousImagePos.Eq(image.Point{-1, -1}) {
		c.swapImageRegion(ctx, s)
	}
	c.previousImagePos = newImagePos
	c.previousSelected = selected
	switch {
	case newImagePos.Y < 0:
		//if the image upper left corner is outside of the screen leave some upper sixel rows
		setCursorPos(w, newImagePos.X, 0)
		leaveRows := int((ctx.YCellPixels()*(-newImagePos.Y))/6) + 4
		c.sixelData.WriteLeaveUpper(w, leaveRows)
	case ctx.YCellPixels()*newImagePos.Y+c.scaledImageBounds.Dy() > int(ctx.YPixel):
		//if the image lover pars is outside of the screen leave some lower sixel rows
		setCursorPos(w, newImagePos.X, newImagePos.Y)
		leaveRows := ((ctx.YCellPixels()*newImagePos.Y+c.scaledImageBounds.Dy())-int(ctx.YPixel))/6 + 2
		c.sixelData.WriteLeaveLower(w, leaveRows)
	default:
		setCursorPos(w, newImagePos.X, newImagePos.Y)
		c.sixelData.Write(w)
	}
}

func (c *Card) swapImageRegion(ctx Context, s tcell.Screen) {
	selected := c.Card == photon.SelectedCard
	style := tcell.StyleDefault
	if selected {
		style = tcell.StyleDefault.Background(selectedColor)
	}
	for x := ctx.X; x < ctx.Width+ctx.X; x++ {
		for y := ctx.Y; y < ctx.Height-headerHeight+ctx.Y; y++ {
			r := '\u2800'
			c, _, _, _ := s.GetContent(x, y)
			if c == r {
				r = '\u2007'
			}
			s.SetContent(x, y, r, nil, style)
		}
	}
}

func (c *Card) DownloadImage(ctx Context, s tcell.Screen) bool {
	if c.ItemImage != nil || c.Item.Image == nil {
		c.makeSixel(ctx, s)
		return false
	}
	photon.ImgDownloader.Download(
		c.Item.Image.URL,
		func(i interface{}) {
			c.ItemImage = i
			c.makeSixel(ctx, s)
		},
	)
	return true
}

func (c *Card) makeSixel(ctx Context, s tcell.Screen) {
	if c.sixelData != nil || c.ItemImage == nil {
		return
	}
	targetWidth := ctx.Width * ctx.XCellPixels()
	targetHeight := (ctx.Height - headerHeight) * ctx.YCellPixels()
	imageProc(
		c,
		c.ItemImage,
		targetWidth,
		targetHeight,
		func(b image.Rectangle, s *Sixel) {
			c.scaledImageBounds, c.sixelData = b, s
			//if c.isOnScreen(c.Card) {
			redraw(false)
			//}
		},
	)
}

func (c *Card) ClearImage() {
	c.sixelData = nil
}
