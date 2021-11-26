package main

import (
	"bytes"
	"fmt"
	"image"
	"io"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/mattn/go-sixel"
	"golang.org/x/image/draw"
)

type Card struct {
	Title         string
	SelectedColor tcell.Color
	Image         image.Image
	selected      bool
	sixelData     []byte
	scaledImage   image.Image
}

func (c *Card) Draw(ctx Context, s tcell.Screen, w io.Writer) {
	if c.selected {
		for x := 0; x < ctx.Width; x++ {
			for y := 0; y < ctx.Height; y++ {
				s.SetContent(x+ctx.X, y+ctx.Y, ' ', nil, tcell.StyleDefault.Background(c.SelectedColor))
			}
		}
	}
	var x int
	for _, c := range c.Title {
		if x > ctx.Width {
			break
		}
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(
			x+ctx.X,
			ctx.Height-1+ctx.Y,
			c,
			comb,
			tcell.StyleDefault,
		)
		x += w
	}
	sw, sh := s.Size()
	targetWidth := ctx.Width * int(ctx.XPixel) / sw
	targetHeight := ctx.Height * int(ctx.YPixel) / sh
	if c.sixelData == nil && c.Image != nil {
		c.scaledImage = resizeImage(c.Image, uint(targetWidth), uint(targetHeight))
		var buf bytes.Buffer
		sixel.NewEncoder(&buf).Encode(c.scaledImage)
		c.sixelData = buf.Bytes()
	}
	imgHeight := c.scaledImage.Bounds().Dy()
	if int(ctx.YPixel)/sh*(ctx.Y+1)+imgHeight > int(ctx.YPixel) {
		return
	}
	if ctx.Y+1 < 0 {
		return
	}
	fmt.Fprintf(w, "\033[%d;%dH", ctx.Y+1, ctx.X+1)
	w.Write(c.sixelData)
}

func (c *Card) ClearImage() {
	c.sixelData = nil
}

func (c *Card) Selected() bool {
	if c == nil {
		return false
	}
	return c.selected
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

func resizeImage(src image.Image, maxWidth, maxHeight uint) image.Image {
	origBounds := src.Bounds()
	origWidth := uint(origBounds.Dx())
	origHeight := uint(origBounds.Dy())
	newWidth, newHeight := origWidth, origHeight

	// Preserve aspect ratio
	if origWidth > maxWidth {
		newHeight = uint(origHeight * maxWidth / origWidth)
		if newHeight < 1 {
			newHeight = 1
		}
		newWidth = maxWidth
	}

	if newHeight > maxHeight {
		newWidth = uint(newWidth * maxHeight / newHeight)
		if newWidth < 1 {
			newWidth = 1
		}
		newHeight = maxHeight
	}
	rect := image.Rect(0, 0, int(newWidth), int(newHeight))
	dst := image.NewRGBA(rect)
	draw.NearestNeighbor.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}
