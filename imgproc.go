package main

import (
	"bytes"
	"image"
	"runtime"
	"sync"

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
)

func init() {
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
		draw.ApproxBiLinear.Scale(dst, rect, req.src, origBounds, draw.Over, nil)
		var buf bytes.Buffer
		sixel.NewEncoder(&buf).Encode(dst)
		req.callback(dst.Bounds(), buf.Bytes())
		//imageProcMap.Delete(req.card)
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
