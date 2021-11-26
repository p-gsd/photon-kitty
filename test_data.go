package main

import (
	"fmt"
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
	c := tcell.NewHexColor(int32(rand.Int() & 0xffffff))

	grid.Children = []Child{}

	for i := 0; i < 100; i++ {
		grid.Children = append(
			grid.Children,
			&Card{
				Title:         fmt.Sprintf("child: %d", i),
				SelectedColor: c,
				Image:         img,
			},
		)
	}
}
