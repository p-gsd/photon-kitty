package main

import (
	"fmt"
	"image"
	"io"

	"git.sr.ht/~ghost08/libphoton"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var cards = make(map[*libphoton.Card]*Card)

func getCard(card *libphoton.Card) *Card {
	c, ok := cards[card]
	if !ok {
		c = &Card{
			Card:          card,
			SelectedColor: tcell.ColorGrey,
		}
		cards[card] = c
	}
	return c
}

type Card struct {
	*libphoton.Card
	SelectedColor     tcell.Color
	selected          bool
	sixelData         []byte
	scaledImageBounds image.Rectangle
}

func drawLines(s tcell.Screen, X, Y, maxWidth, maxLines int, text string, style tcell.Style) {
	var x, y int
	for _, c := range text {
		if x > maxWidth {
			y++
			x = 0
			if y >= maxLines {
				break
			}
		}
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x+X, y+Y, c, comb, style)
		x += w
	}
}

func (c *Card) Draw(ctx Context, s tcell.Screen, w io.Writer) {
	background := tcell.ColorBlack
	if c.selected {
		background = c.SelectedColor
		for x := 0; x < ctx.Width; x++ {
			for y := 0; y < ctx.Height; y++ {
				s.SetContent(x+ctx.X, y+ctx.Y, ' ', nil, tcell.StyleDefault.Background(c.SelectedColor))
			}
		}
	}
	s.SetContent(ctx.X, ctx.Height-2+ctx.Y, ' ', nil, tcell.StyleDefault.Background(background))
	s.SetContent(ctx.X, ctx.Height-1+ctx.Y, ' ', nil, tcell.StyleDefault.Background(background))
	drawLines(
		s,
		ctx.X+1,
		ctx.Height-2+ctx.Y,
		ctx.Width-3,
		2,
		c.Item.Title,
		tcell.StyleDefault.Background(background),
	)
	if c.DownloadImage(ctx, s) {
		return
	}
	c.makeSixel(ctx, s)
	if c.sixelData == nil {
		return
	}
	imgHeight := c.scaledImageBounds.Dy()
	if int(ctx.YPixel)/int(ctx.Rows)*(ctx.Y+1)+imgHeight > int(ctx.YPixel) {
		return
	}
	if ctx.Y+1 < 0 {
		return
	}
	fmt.Fprintf(w, "\033[%d;%dH", ctx.Y+1, ctx.X+1)
	w.Write(c.sixelData)
}

func (c *Card) DownloadImage(ctx Context, s tcell.Screen) bool {
	if c.ItemImage != nil || c.Item.Image == nil {
		return false
	}
	photon.ImgDownloader.Download(
		c.Item.Image.URL,
		func(img image.Image) {
			c.ItemImage = img
			c.makeSixel(ctx, s)
			redraw(false)
		},
	)
	return true
}

func (c *Card) makeSixel(ctx Context, s tcell.Screen) {
	if c.sixelData == nil && c.ItemImage != nil {
		targetWidth := ctx.Width * int(ctx.XPixel) / int(ctx.Cols)
		targetHeight := (ctx.Height - 2) * int(ctx.YPixel) / int(ctx.Rows)
		c.scaledImageBounds, c.sixelData = imageProc(c.ItemImage, targetWidth, targetHeight)
	}
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
