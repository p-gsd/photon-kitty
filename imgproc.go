package main

import (
	"bytes"
	"image"
	"image/color"
	"runtime"
	"sync"
	"time"

	"github.com/afbjorklund/go-termbg/pkg/termbg"
	"github.com/mattn/go-sixel"
	"golang.org/x/image/draw"
)

//Image Processing, scaling and sixel encoding

type imageProcReq struct {
	card      *Card
	src       image.Image
	maxWidth  int
	maxHeight int
	callback  func(image.Rectangle, []byte)
}

var (
	imageProcChan = make(chan imageProcReq, 1024)
	//map that holds cards that are right now processed
	imageProcMap sync.Map
	terminalBg   color.Color
)

func init() {
	if rgb, err := termbg.NewRGB(100 * time.Millisecond); err == nil {
		terminalBg = color.RGBA{
			R: uint8(rgb.R >> 8),
			G: uint8(rgb.G >> 8),
			B: uint8(rgb.B >> 8),
			A: 255,
		}
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		go imageProcWorker()
	}
}

func imageProcWorker() {
	for req := range imageProcChan {
		//if there is a goroutine that already processes this card, then skip
		if _, ok := imageProcMap.LoadOrStore(req.card, struct{}{}); ok {
			continue
		}
		origBounds := req.src.Bounds()
		origWidth := origBounds.Dx()
		origHeight := origBounds.Dy()
		newWidth, newHeight := origWidth, origHeight
		// Preserve aspect ratio
		if origWidth > origHeight {
			newHeight = origHeight * req.maxWidth / origWidth
			newWidth = req.maxWidth
		} else {
			newWidth = origWidth * req.maxHeight / origHeight
			newHeight = req.maxHeight
		}
		offset := (req.maxWidth - newWidth) / 2
		rect := image.Rect(offset, 0, newWidth+offset, newHeight)
		dst := image.NewRGBA(image.Rect(0, 0, req.maxWidth, req.maxHeight))
		draw.Draw(dst, dst.Bounds(), &image.Uniform{terminalBg}, image.Point{}, draw.Src)
		draw.ApproxBiLinear.Scale(dst, rect, req.src, origBounds, draw.Over, nil)
		var buf bytes.Buffer
		sixel.NewEncoder(&buf).Encode(dst)
		req.callback(dst.Bounds(), buf.Bytes())
	}
}

func imageProc(card *Card, src image.Image, maxWidth, maxHeight int, callback func(image.Rectangle, []byte)) {
	imageProcChan <- imageProcReq{
		card:      card,
		src:       src,
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
		callback:  callback,
	}
}

func imageProcClear() {
	imageProcMap.Range(func(k, v interface{}) bool {
		imageProcMap.Delete(k)
		return true
	})
}
