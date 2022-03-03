//go:build windows

package imgproc

import (
	"fmt"
	"image"
)

func Init() error {
	return fmt.Errorf("no windows support")
}

func NewImageResizer(img image.Image) (ImageResizer, error) {
	return nil, fmt.Errorf("no windows support")
}
