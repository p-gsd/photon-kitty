package imgproc

import (
	"image"

	"github.com/soniakeys/quant/median"
	"golang.org/x/image/draw"
)

type CPUImageResizer struct {
	img image.Image
}

func (cir *CPUImageResizer) Release() error { return nil }

func (cir *CPUImageResizer) Resize(maxWidth, maxHeight uint) (image.Image, error) {

	origBounds := cir.img.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()
	newWidth, newHeight := outSize(
		uint(origWidth),
		uint(origHeight),
		uint(maxWidth),
		uint(maxHeight),
	)
	rect := image.Rect(0, 0, int(newWidth), int(newHeight))
	dst := image.NewRGBA(rect)
	draw.NearestNeighbor.Scale(dst, rect, cir.img, origBounds, draw.Over, nil)
	return dst, nil
}

func (cir *CPUImageResizer) ResizePaletted(p, maxWidth, maxHeight uint) (*image.Paletted, error) {
	img, err := cir.Resize(maxWidth, maxHeight)
	if err != nil {
		return nil, err
	}
	if p, ok := img.(*image.Paletted); ok {
		return p, nil
	}
	// make adaptive palette using median cut alogrithm
	q := median.Quantizer(p - 1)
	paletted := q.Paletted(img)
	draw.Draw(paletted, img.Bounds(), img, image.Point{}, draw.Over)
	return paletted, nil
}
