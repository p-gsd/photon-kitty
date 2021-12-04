package main

import (
	"fmt"
	"image"
	"io"
	"strings"
	"time"

	"git.sr.ht/~ghost08/libphoton"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	htime "github.com/sbani/go-humanizer/time"
)

var cards = make(map[*libphoton.Card]*Card)

func getCard(card *libphoton.Card) *Card {
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
	*libphoton.Card
	selected          bool
	sixelData         []byte
	scaledImageBounds image.Rectangle
	//isOnScreen        func(*libphoton.Card)
	previousImagePos image.Point
	previousSelected bool
}

func drawLine(s tcell.Screen, X, Y, maxWidth int, text string, style tcell.Style) (width int) {
	var x int
	for _, c := range text {
		if c == '\n' {
			return
		}
		if x > maxWidth {
			return
		}
		var comb []rune
		w := runewidth.RuneWidth(c)
		width += w
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}

		s.SetContent(x+X, Y, c, comb, style)
		x += w
	}
	return
}

func drawLinesWordwrap(s tcell.Screen, X, Y, maxWidth, maxLines int, text string, style tcell.Style) {
	var x, y int
	var word strings.Builder
	var wordLength int
	for _, c := range text {
		if c != ' ' && c != '\n' {
			word.WriteRune(c)
			wordLength += runewidth.RuneWidth(c)
			continue
		}
		for wordLength > maxWidth && y < maxLines {
			w := drawLine(s, x+X, y+Y, maxWidth-x, word.String(), style)
			wordRest := word.String()[w:]
			word.Reset()
			word.WriteString(wordRest)
			wordLength -= w
			y++
			x = 0
		}
		if y >= maxLines {
			break
		}
		if c == '\n' || x+wordLength == maxWidth {
			drawString(s, x+X, y+Y, word.String(), style)
			word.Reset()
			wordLength = 0
			y++
			x = 0
			continue
		}
		if x+wordLength > maxWidth {
			y++
			x = 0
		}
		if y >= maxLines {
			break
		}
		x += drawString(s, x+X, y+Y, word.String()+" ", style)
		word.Reset()
		wordLength = 0
	}
}

func drawString(s tcell.Screen, x, y int, text string, style tcell.Style) (width int) {
	for _, c := range text {
		var comb []rune
		w := runewidth.RuneWidth(c)
		width += w
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}

		s.SetContent(x, y, c, comb, style)
		x += w
	}
	return
}

func (c *Card) Draw(ctx Context, s tcell.Screen, w io.Writer) {
	background := tcell.ColorBlack
	if c.selected {
		background = selectedColor
	}
	if c.Item.Image == nil {
		for x := ctx.X; x < ctx.Width+ctx.X; x++ {
			for y := ctx.Y; y < ctx.Height+ctx.Y; y++ {
				s.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(background))
			}
		}
		drawLinesWordwrap(s, ctx.X+1, ctx.Y, ctx.Width-3, 2, c.Item.Title, tcell.StyleDefault.Background(background).Bold(true))
		drawLine(s, ctx.X+1, ctx.Y+2, ctx.Width-3, c.Feed.Title, tcell.StyleDefault.Background(background).Italic(true))
		drawLine(s, ctx.X+1, ctx.Y+3, ctx.Width-3, htime.Difference(time.Now(), *c.Item.PublishedParsed), tcell.StyleDefault.Background(background).Italic(true))
		drawLinesWordwrap(s, ctx.X+1, ctx.Y+headerHeight+1, ctx.Width-3, ctx.Height-headerHeight-2, c.Item.Description, tcell.StyleDefault.Background(background))
		return
	}

	//header
	for x := ctx.X; x < ctx.Width+ctx.X; x++ {
		for y := ctx.Height - headerHeight + ctx.Y; y < ctx.Height+ctx.Y; y++ {
			s.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(background))
		}
	}
	drawLinesWordwrap(s, ctx.X+1, ctx.Height-headerHeight+ctx.Y, ctx.Width-3, 2, c.Item.Title, tcell.StyleDefault.Background(background).Bold(true))
	drawLine(s, ctx.X+1, ctx.Height-headerHeight+ctx.Y+2, ctx.Width-3, c.Feed.Title, tcell.StyleDefault.Background(background).Italic(true))
	drawLine(s, ctx.X+1, ctx.Height-headerHeight+ctx.Y+3, ctx.Width-3, htime.Difference(time.Now(), *c.Item.PublishedParsed), tcell.StyleDefault.Background(background).Italic(true))

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
	imgHeight := c.scaledImageBounds.Dy()
	if ctx.YCellPixels()*(ctx.Y+1)+imgHeight > int(ctx.YPixel) {
		if !c.previousImagePos.Eq(image.Point{-2, -2}) {
			c.previousImagePos = image.Point{-2, -2}
			c.swapImageRegion(ctx, s)
		}
		return
	}
	if ctx.Y+1 < 0 {
		if !c.previousImagePos.Eq(image.Point{-2, -2}) {
			c.previousImagePos = image.Point{-2, -2}
			c.swapImageRegion(ctx, s)
		}
		return
	}
	imageWidthInCells := c.scaledImageBounds.Dx() / ctx.XCellPixels()
	offset := (ctx.Width - imageWidthInCells) / 2
	newImagePos := image.Point{ctx.X + 1 + offset, ctx.Y + 1}
	if c.previousImagePos.Eq(newImagePos) && c.selected == c.previousSelected {
		return
	}
	if !c.previousImagePos.Eq(image.Point{-1, -1}) {
		c.swapImageRegion(ctx, s)
	}
	c.previousImagePos = newImagePos
	c.previousSelected = c.selected
	fmt.Fprintf(w, "\033[%d;%dH", newImagePos.Y, newImagePos.X) //set cursor to x, y
	w.Write(c.sixelData)
}

func (c *Card) swapImageRegion(ctx Context, s tcell.Screen) {
	background := tcell.ColorBlack
	if c.selected {
		background = selectedColor
	}
	for x := ctx.X; x < ctx.Width+ctx.X; x++ {
		for y := ctx.Y; y < ctx.Height-headerHeight+ctx.Y; y++ {
			r := '\u2800'
			c, _, _, _ := s.GetContent(x, y)
			if c == r {
				r = '\u2007'
			}
			s.SetContent(x, y, r, nil, tcell.StyleDefault.Background(background))
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
		func(img image.Image) {
			c.ItemImage = img
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
		func(b image.Rectangle, sd []byte) {
			c.scaledImageBounds, c.sixelData = b, sd
			//if c.isOnScreen(c.Card) {
			redraw(false)
			//}
		},
	)
}

func (c *Card) ClearImage() {
	c.sixelData = nil
}

func (c *Card) Select() {
	if c == nil {
		return
	}
	c.selected = true
}

func (c *Card) Unselect() {
	if c == nil {
		return
	}
	c.selected = false
}
