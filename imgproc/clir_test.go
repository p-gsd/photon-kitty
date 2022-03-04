package imgproc

import (
	"image"
	"image/jpeg"
	"os"
	"sync"
	"testing"
)

func TestResize(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("yoda.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := NewOpenCLImageResizer(img)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		img1, err := ir.Resize(500, 500)
		if err != nil {
			panic(err)
		}
		f1, err := os.Create("yoda_1.jpg")
		if err != nil {
			panic(err)
		}
		defer f1.Close()
		if err := jpeg.Encode(f1, img1, nil); err != nil {
			panic(err)
		}
	}()

	img2, err := ir.Resize(200, 200)
	if err != nil {
		t.Fatal(err)
	}
	f2, err := os.Create("yoda_2.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()
	if err := jpeg.Encode(f2, img2, nil); err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		img3, err := ir.Resize(100, 100)
		if err != nil {
			panic(err)
		}
		f3, err := os.Create("yoda_3.jpg")
		if err != nil {
			panic(err)
		}
		defer f3.Close()
		if err := jpeg.Encode(f3, img3, nil); err != nil {
			panic(err)
		}
	}()

	ir2, err := NewOpenCLImageResizer(img)
	if err != nil {
		t.Fatal(err)
	}
	img4, err := ir2.Resize(200, 200)
	if err != nil {
		t.Fatal(err)
	}
	f4, err := os.Create("yoda_4.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f4.Close()
	if err := jpeg.Encode(f4, img4, nil); err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	if err := ir.Release(); err != nil {
		t.Fatal(err)
	}
	if err := ir2.Release(); err != nil {
		t.Fatal(err)
	}
}

func TestResizePaletted(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("yoda.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := NewOpenCLImageResizer(img)
	if err != nil {
		t.Fatal(err)
	}
	pImg, err := ir.ResizePaletted(255, 1000, 600)
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.Create("yoda_paletted.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, pImg, nil); err != nil {
		t.Fatal(err)
	}
	if err := ir.Release(); err != nil {
		t.Fatal(err)
	}
}
