package main

import (
	"image"
	"log"
	"runtime"
	"sync"

	"git.sr.ht/~ghost08/clir"
	"golang.org/x/image/draw"
)

//Image Processing, scaling and sixel encoding

type imageProcReq struct {
	ident     interface{}
	src       interface{}
	maxWidth  int
	maxHeight int
	callback  func(image.Rectangle, *Sixel)
}

const numColors = 255

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
		if _, ok := imageProcMap.LoadOrStore(req.ident, struct{}{}); ok {
			continue
		}
		img := resize(req.ident, req.src, req.maxWidth, req.maxHeight)
		if img == nil {
			continue
		}
		sixel := EncodeSixel(numColors, img)
		if !imageProcStillThere(req.ident) {
			continue
		}
		req.callback(img.Bounds(), sixel)
	}
}

func resize(ident, obj interface{}, maxWidth, maxHeight int) image.Image {
	switch i := obj.(type) {
	case image.Image:
		origBounds := i.Bounds()
		origWidth := origBounds.Dx()
		origHeight := origBounds.Dy()
		newWidth, newHeight := origWidth, origHeight
		// Preserve aspect ratio
		if origWidth >= origHeight {
			newHeight = origHeight * maxWidth / origWidth
			newWidth = maxWidth
		} else {
			newWidth = origWidth * maxHeight / origHeight
			newHeight = maxHeight
		}
		if newHeight > maxHeight {
			newHeight = maxHeight
			newWidth = origWidth * maxHeight / origHeight
		}
		rect := image.Rect(0, 0, newWidth, newHeight)
		dst := image.NewRGBA(rect)
		if !imageProcStillThere(ident) {
			return nil
		}
		draw.NearestNeighbor.Scale(dst, rect, i, origBounds, draw.Over, nil)
		if !imageProcStillThere(ident) {
			return nil
		}
		return dst
	case *clir.ImageResizer:
		img, err := i.ResizePaletted(numColors-1, uint(maxWidth), uint(maxHeight))
		if err != nil {
			log.Printf("ERROR: opencl image resizer error, falling back to CPU image scaling: %v", err)
			imageCache.gotError = true
			return nil
		}
		return img
	default:
		log.Panicf("scale image got %T", i)
		return nil
	}
}

func imageProc(
	ident interface{},
	src interface{},
	maxWidth,
	maxHeight int,
	callback func(image.Rectangle, *Sixel),
) {
	imageProcChan <- imageProcReq{
		ident:     ident,
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
func imageProcStillThere(ident interface{}) bool {
	_, ok := imageProcMap.Load(ident)
	return ok
}

type ClirCache struct {
	m        sync.Map
	gotError bool
}

func (cc *ClirCache) Load(key interface{}) (interface{}, bool) {
	return cc.m.Load(key)
}

func (cc *ClirCache) Store(key interface{}, val interface{}) {
	switch i := val.(type) {
	case image.Image:
		if cc.gotError {
			cc.m.Store(key, val)
			break
		}
		v, err := clir.NewImageResizer(i)
		if err != nil {
			log.Println("ERROR: opencl image resizer, falling back to CPU scaling: %w", err)
			cc.gotError = true
			cc.m.Store(key, val)
			break
		}
		cc.m.Store(key, v)
	case *clir.ImageResizer:
		cc.m.Store(key, i)
	case nil:
		cc.m.Store(key, nil)
		return
	default:
		log.Panicf("ERROR: ClirCache got val type %T", val)
	}
}
