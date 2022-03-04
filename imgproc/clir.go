package imgproc

import (
	"image"
	"image/color"
	"log"
)

//global variable that stores if loading opencl returned an error
var gotError bool

type ImageResizer interface {
	Release() error
	Resize(maxWidth, maxHeight uint) (image.Image, error)
	ResizePaletted(p, maxWidth, maxHeight uint) (*image.Paletted, error)
}

func NewImageResizer(i interface{}) ImageResizer {
	if ir, ok := i.(ImageResizer); ok {
		return ir
	}
	img := i.(image.Image)
	if gotError {
		return &CPUImageResizer{img: img}
	}
	ir, err := NewOpenCLImageResizer(img)
	if err != nil {
		gotError = true
		log.Println("ERROR: opencl image resizer:", err)
		return &CPUImageResizer{img: img}
	}
	return ir
}

func outSize(origWidth, origHeight, maxWidth, maxHeight uint) (uint, uint) {
	outWidth, outHeight := origWidth, origHeight
	// Preserve aspect ratio
	if origWidth >= origHeight {
		outHeight = origHeight * maxWidth / origWidth
		outWidth = maxWidth
	} else {
		outWidth = origWidth * maxHeight / origHeight
		outHeight = maxHeight
	}
	if outHeight > maxHeight {
		outHeight = maxHeight
		outWidth = origWidth * maxHeight / origHeight
	}
	return outWidth, outHeight
}

func imgData(i image.Image) []byte {
	switch m := i.(type) {
	case *image.Gray:
		return m.Pix
	case *image.RGBA:
		return m.Pix
	}

	b := i.Bounds()
	w := b.Dx()
	h := b.Dy()
	data := make([]byte, w*h*4)
	dataOffset := 0
	colorChan := make(chan color.Color)
	go func() {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				colorChan <- i.At(x+b.Min.X, y+b.Min.Y)
			}
		}
		close(colorChan)
	}()
	for c := range colorChan {
		r, g, b, a := c.RGBA()
		data[dataOffset] = uint8(r >> 8)
		data[dataOffset+1] = uint8(g >> 8)
		data[dataOffset+2] = uint8(b >> 8)
		data[dataOffset+3] = uint8(a >> 8)
		dataOffset += 4
	}
	return data
}
