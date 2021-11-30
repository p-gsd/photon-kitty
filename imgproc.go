package main

import (
	"bytes"
	"image"
	"runtime"

	"github.com/mattn/go-sixel"
	"golang.org/x/image/draw"
)

//Image Processing, scaling and sixel encoding

type imageProcReq struct {
	src       image.Image
	maxWidth  int
	maxHeight int
	resp      chan imageProcResp
}

type imageProcResp struct {
	scaledImageBounds image.Rectangle
	sixelData         []byte
}

var imageProcChan = make(chan imageProcReq, 50)

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
		draw.ApproxBiLinear.Scale(dst, rect, req.src, origBounds, draw.Over, nil)
		var buf bytes.Buffer
		sixel.NewEncoder(&buf).Encode(dst)
		req.resp <- imageProcResp{
			scaledImageBounds: dst.Bounds(),
			sixelData:         buf.Bytes(),
		}
	}
}

func imageProc(src image.Image, maxWidth, maxHeight int) (image.Rectangle, []byte) {
	respChan := make(chan imageProcResp)
	imageProcChan <- imageProcReq{
		src:       src,
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
		resp:      respChan,
	}
	resp := <-respChan
	return resp.scaledImageBounds, resp.sixelData
}
