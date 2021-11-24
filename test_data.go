package main

import (
	"image/png"
	"math/rand"
	"os"

	"github.com/gdamore/tcell/v2"
)

func addTestDataToGrid(grid *Grid) {
	f, err := os.Open("yoda.png")
	if err != nil {
		panic(err)
	}
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	grid.Children = []Child{
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
		&Card{Title: "Foobar", Color: tcell.NewHexColor(int32(rand.Int() & 0xffffff)), Image: img},
	}

}
