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
		if origWidth >= origHeight {
			newHeight = origHeight * req.maxWidth / origWidth
			newWidth = req.maxWidth
		} else {
			newWidth = origWidth * req.maxHeight / origHeight
			newHeight = req.maxHeight
		}
		if newHeight > req.maxHeight {
			newHeight = req.maxHeight
			newWidth = origWidth * req.maxHeight / origHeight
		}
		rect := image.Rect(0, 0, newWidth, newHeight)
		dst := image.NewRGBA(rect)
		if !imageProcStillThere(req.card) {
			continue
		}
		draw.ApproxBiLinear.Scale(dst, rect, req.src, origBounds, draw.Over, nil)
		if !imageProcStillThere(req.card) {
			continue
		}
		var buf bytes.Buffer
		sixel.NewEncoder(&buf).Encode(dst)
		if !imageProcStillThere(req.card) {
			continue
		}
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

//clears the image map, serves for EventResize
func imageProcClear() {
	imageProcMap.Range(func(k, v interface{}) bool {
		imageProcMap.Delete(k)
		return true
	})
}

//checks if image is still in map, server as checking if
//the map wasn't cleared and all images must be rescaled
func imageProcStillThere(c *Card) bool {
	_, ok := imageProcMap.Load(c)
	return ok
}
