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
	previousImagePos  image.Point
	previousSelected  bool
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
	if c.Item.Image == nil {
		for x := 0; x < ctx.Width; x++ {
			for y := 0; y < ctx.Height; y++ {
				s.SetContent(x+ctx.X, y+ctx.Y, ' ', nil, tcell.StyleDefault.Background(background))
			}
		}
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
	for x := 0; x < ctx.Width; x++ {
		for y := 0; y < ctx.Height-headerHeight; y++ {
			s.SetContent(x+ctx.X, y+ctx.Y, '\u2800', nil, tcell.StyleDefault.Background(background))
		}
	}
	var imageDrawn bool
	defer func() {
		if !imageDrawn {
			for x := 0; x < ctx.Width; x++ {
				for y := 0; y < ctx.Height-headerHeight; y++ {
					s.SetContent(x+ctx.X, y+ctx.Y, ' ', nil, tcell.StyleDefault.Background(background))
				}
			}
		}
		for x := 0; x < ctx.Width; x++ {
			for y := ctx.Height - headerHeight; y < ctx.Height; y++ {
				s.SetContent(x+ctx.X, y+ctx.Y, ' ', nil, tcell.StyleDefault.Background(background))
			}
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
	}()
	if c.DownloadImage(ctx, s) {
		c.previousImagePos = image.Point{-1, -1}
		return
	}
	if c.sixelData == nil {
		c.previousImagePos = image.Point{-1, -1}
		return
	}
	imgHeight := c.scaledImageBounds.Dy()
	if int(ctx.YPixel)/int(ctx.Rows)*(ctx.Y+1)+imgHeight > int(ctx.YPixel) {
		c.previousImagePos = image.Point{-1, -1}
		return
	}
	if ctx.Y+1 < 0 {
		c.previousImagePos = image.Point{-1, -1}
		return
	}
	imageDrawn = true
	newImagePos := image.Point{ctx.X + 1, ctx.Y + 1}
	if c.previousImagePos.Eq(newImagePos) && c.selected == c.previousSelected {
		return
	}
	c.previousImagePos = newImagePos
	c.previousSelected = c.selected
	//set cursor to x, y
	fmt.Fprintf(w, "\033[%d;%dH", newImagePos.Y, newImagePos.X)
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
