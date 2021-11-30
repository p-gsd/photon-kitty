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
			Card: card,
		}
		cards[card] = c
	}
	return c
}

const (
	headerHeight  = 2
	selectedColor = tcell.ColorGray
)

type Card struct {
	*libphoton.Card
	selected          bool
	sixelData         []byte
	scaledImageBounds image.Rectangle
}

func drawLines(s tcell.Screen, X, Y, maxWidth, maxLines int, text string, style tcell.Style) {
	var x, y int
	for _, c := range text {
		if c == '\n' {
			y++
			x = 0
			continue
		}
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
		background = selectedColor
	}
	for x := 0; x < ctx.Width; x++ {
		for y := 0; y < ctx.Height; y++ {
			s.SetContent(x+ctx.X, y+ctx.Y, ' ', nil, tcell.StyleDefault.Background(background))
		}
	}
	if c.Item.Image == nil {
		drawLines(
			s,
			ctx.X+1,
			ctx.Y,
			ctx.Width-3,
			headerHeight,
			c.Item.Title,
			tcell.StyleDefault.Background(background).Bold(true),
		)
		drawLines(
			s,
			ctx.X+1,
			ctx.Y+headerHeight,
			ctx.Width-3,
			ctx.Height-headerHeight,
			c.Item.Description,
			tcell.StyleDefault.Background(background),
		)
		return
	}
	drawLines(
		s,
		ctx.X+1,
		ctx.Height-headerHeight+ctx.Y,
		ctx.Width-3,
		headerHeight,
		c.Item.Title,
		tcell.StyleDefault.Background(background).Bold(true),
	)
	if c.DownloadImage(ctx, s) {
		return
	}
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
	targetWidth := ctx.Width * int(ctx.XPixel) / int(ctx.Cols)
	targetHeight := (ctx.Height - headerHeight) * int(ctx.YPixel) / int(ctx.Rows)
	imageProc(
		c,
		c.ItemImage,
		targetWidth,
		targetHeight,
		func(b image.Rectangle, sd []byte) {
			c.scaledImageBounds, c.sixelData = b, sd
			redraw(false)
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
