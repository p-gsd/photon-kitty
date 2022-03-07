//go:build linux

package imgproc

/*
#cgo LDFLAGS: -ldl

int init();
void* createImageResizer(unsigned int,unsigned int,unsigned int,unsigned int,void*,int*);
int releaseImageResizer(void*);
int resize(void*,unsigned int,unsigned int,void*);
int resize_paletted(void*,unsigned int,unsigned int,void*,void*,unsigned int);
*/
import "C"
import (
	"fmt"
	"image"
	"image/color"
	"unsafe"
)

func Init() error {
	ret := C.init()
	if ret != C.int(0) {
		gotError = true
		return fmt.Errorf("init error %d", ret)
	}
	return nil
}

type ImageResizerOpenCL struct {
	cir       unsafe.Pointer
	bounds    image.Rectangle
	pixelSize uint
}

func NewOpenCLImageResizer(img image.Image) (ImageResizer, error) {
	ir := &ImageResizerOpenCL{bounds: img.Bounds()}
	var rowPitch uint
	if g, ok := img.(*image.Gray); ok {
		rowPitch = uint(g.Stride)
		ir.pixelSize = 1
	} else {
		ir.pixelSize = 4
	}
	data := imgData(img)
	var ret C.int
	ir.cir = C.createImageResizer(
		C.uint(img.Bounds().Dx()),
		C.uint(img.Bounds().Dy()),
		C.uint(rowPitch),
		C.uint(ir.pixelSize),
		unsafe.Pointer(&data[0]),
		&ret,
	)
	if ret != C.int(0) {
		return nil, fmt.Errorf("opencl error: %d", ret)
	}
	return ir, nil
}

func (ir *ImageResizerOpenCL) Release() error {
	if ret := C.releaseImageResizer(ir.cir); ret != C.int(0) {
		return fmt.Errorf("opencl error: %d", ret)
	}
	return nil
}

func (ir *ImageResizerOpenCL) Resize(maxWidth, maxHeight uint) (image.Image, error) {
	origWidth := uint(ir.bounds.Dx())
	origHeight := uint(ir.bounds.Dy())
	outWidth, outHeight := outSize(origWidth, origHeight, maxWidth, maxHeight)

	outData := make([]byte, outWidth*outHeight*ir.pixelSize)
	ret := C.resize(
		ir.cir,
		C.uint(outWidth),
		C.uint(outHeight),
		unsafe.Pointer(&outData[0]),
	)
	if ret != C.int(0) {
		return nil, fmt.Errorf("opencl error: %d", ret)
	}

	if ir.pixelSize == 1 {
		out := image.NewGray(image.Rect(0, 0, int(outWidth), int(outHeight)))
		out.Pix = outData
		return out, nil
	}
	out := image.NewRGBA(image.Rect(0, 0, int(outWidth), int(outHeight)))
	out.Pix = outData
	return out, nil
}

func (ir *ImageResizerOpenCL) ResizePaletted(p, maxWidth, maxHeight uint) (*image.Paletted, error) {
	if maxWidth == 0 || maxHeight == 0 {
		return nil, nil
	}
	origWidth := uint(ir.bounds.Dx())
	origHeight := uint(ir.bounds.Dy())
	outWidth, outHeight := outSize(origWidth, origHeight, maxWidth, maxHeight)

	outData := make([]byte, outWidth*outHeight)
	paletteData := make([]uint8, p*4)
	ret := C.resize_paletted(
		ir.cir,
		C.uint(outWidth),
		C.uint(outHeight),
		unsafe.Pointer(&outData[0]),
		unsafe.Pointer(&paletteData[0]),
		C.uint(p),
	)
	if ret != C.int(0) {
		return nil, fmt.Errorf("opencl error: %d", ret)
	}
	palette := make([]color.Color, p)
	for i := 0; i < int(p); i++ {
		palette[i] = color.RGBA{
			R: paletteData[i*4],
			G: paletteData[i*4+1],
			B: paletteData[i*4+2],
			A: paletteData[i*4+3],
		}
	}
	img := &image.Paletted{
		Pix:     outData,
		Stride:  int(outWidth),
		Rect:    image.Rect(0, 0, int(outWidth), int(outHeight)),
		Palette: palette,
	}
	return img, nil
}
