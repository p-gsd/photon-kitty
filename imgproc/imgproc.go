package imgproc

import (
	"image"
	"log"
	"runtime"
	"sync"
)

//Image Processing, scaling and sixel encoding

type imageProcReq struct {
	ident     interface{}
	src       ImageResizer
	maxWidth  int
	maxHeight int
	callback  func(*Sixel)
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
		img, err := req.src.ResizePaletted(numColors-1, uint(req.maxWidth), uint(req.maxHeight))
		if err != nil {
			log.Printf("ERROR: opencl image resizer error, falling back to CPU image scaling: %v", err)
			gotError = true
		}
		if img == nil {
			continue
		}
		req.callback(EncodeSixel(numColors, img))
	}
}

//sends a image processing request to the workers
func Proc(
	ident interface{},
	src ImageResizer,
	maxWidth,
	maxHeight int,
	callback func(*Sixel),
) {
	imageProcChan <- imageProcReq{
		ident:     ident,
		src:       src,
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
		callback:  callback,
	}
}

func ProcDelete(key interface{}) {
	imageProcMap.Delete(key)
}

//clears the image map, serves for EventResize
func ProcClear() {
	imageProcMap.Range(func(k, v interface{}) bool {
		imageProcMap.Delete(k)
		return true
	})
}

type Cache struct {
	m sync.Map
}

func (cc *Cache) Load(key interface{}) (interface{}, bool) {
	return cc.m.Load(key)
}

func (cc *Cache) Store(key interface{}, val interface{}) {
	switch i := val.(type) {
	case image.Image:
		if gotError {
			cc.m.Store(key, val)
			break
		}
		v, err := NewOpenCLImageResizer(i)
		if err != nil {
			log.Println("ERROR: opencl image resizer, falling back to CPU scaling: %w", err)
			gotError = true
			cc.m.Store(key, CPUImageResizer{img: i})
			break
		}
		cc.m.Store(key, v)
	case *ImageResizer:
		cc.m.Store(key, i)
	case nil:
		cc.m.Store(key, nil)
		return
	default:
		log.Panicf("ERROR: ClirCache got val type %T", val)
	}
}
