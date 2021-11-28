package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"runtime"

	"git.sr.ht/~ghost08/libphoton"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/mattn/go-sixel"
	"golang.org/x/image/draw"
)

var cards = make(map[*libphoton.Card]*Card)

func getCard(card *libphoton.Card) *Card {
	c, ok := cards[card]
	if !ok {
		c = &Card{
			Card:          card,
			SelectedColor: tcell.ColorPurple,
		}
		cards[card] = c
	}
	return c
}

type Card struct {
	*libphoton.Card
	SelectedColor tcell.Color
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
	for _, c := range c.Item.Title {
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
	if c.ItemImage == nil && c.Item.Image != nil {
		photon.ImgDownloader.Download(
			c.Item.Image.URL,
			func(img image.Image) {
				c.ItemImage = img
				c.makeSixel(ctx, s)
				redraw()
			},
		)
		return
	}
	c.makeSixel(ctx, s)
	if c.scaledImage == nil {
		return
	}
	imgHeight := c.scaledImage.Bounds().Dy()
	if int(ctx.YPixel)/int(ctx.Rows)*(ctx.Y+1)+imgHeight > int(ctx.YPixel) {
		return
	}
	if ctx.Y+1 < 0 {
		return
	}
	fmt.Fprintf(w, "\033[%d;%dH", ctx.Y+1, ctx.X+1)
	w.Write(c.sixelData)
}

func (c *Card) makeSixel(ctx Context, s tcell.Screen) {
	if c.sixelData == nil && c.ItemImage != nil {
		targetWidth := ctx.Width * int(ctx.XPixel) / int(ctx.Cols)
		targetHeight := ctx.Height * int(ctx.YPixel) / int(ctx.Rows)
		c.scaledImage, c.sixelData = imageProc(c.ItemImage, targetWidth, targetHeight)
	}
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

type imageProcReq struct {
	src       image.Image
	maxWidth  int
	maxHeight int
	resp      chan imageProcResp
}

type imageProcResp struct {
	scaledImage image.Image
	sixelData   []byte
}

var imageProcChan = make(chan imageProcReq)

func init() {
	for i := 0; i < runtime.NumCPU(); i++ {
		go imageProcWorker()
	}
}

func imageProcWorker() {
	for req := range imageProcChan {
		origBounds := req.src.Bounds()
		origWidth := origBounds.Dx()
		origHeight := origBounds.Dy()
		newWidth, newHeight := origWidth, origHeight

		// Preserve aspect ratio
		if origWidth > req.maxWidth {
			newHeight = origHeight * req.maxWidth / origWidth
			if newHeight < 1 {
				newHeight = 1
			}
			newWidth = req.maxWidth
		}

		if newHeight > req.maxHeight {
			newWidth = newWidth * req.maxHeight / newHeight
			if newWidth < 1 {
				newWidth = 1
			}
			newHeight = req.maxHeight
		}
		rect := image.Rect(0, 0, int(newWidth), int(newHeight))
		dst := image.NewRGBA(rect)
		draw.ApproxBiLinear.Scale(dst, rect, req.src, req.src.Bounds(), draw.Over, nil)
		var buf bytes.Buffer
		sixel.NewEncoder(&buf).Encode(dst)
		req.resp <- imageProcResp{scaledImage: dst, sixelData: buf.Bytes()}
	}
}

func imageProc(src image.Image, maxWidth, maxHeight int) (image.Image, []byte) {
	respChan := make(chan imageProcResp)
	imageProcChan <- imageProcReq{
		src:       src,
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
		resp:      respChan,
	}
	resp := <-respChan
	return resp.scaledImage, resp.sixelData
}
