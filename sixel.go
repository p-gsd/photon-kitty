package main

import (
	"bytes"
	"fmt"
	"image"
	"io"

	"github.com/soniakeys/quant/median"
	"golang.org/x/image/draw"
)

const (
	specialChNr = byte(0x6d)
	specialChCr = byte(0x64)
)

//Sixel are the bytes of a image encoded in sixel format
//it has the ability to draw just specified rows of the image
type Sixel struct {
	palette []byte
	rows    [][]byte
}

func (s *Sixel) Rows() int {
	return len(s.rows)
}

func (s *Sixel) write(w io.Writer, from, to int) {
	// DECSIXEL Introducer(\033P0;0;8q) + DECGRA ("1;1): Set Raster Attributes
	w.Write([]byte{0x1b, 0x50, 0x30, 0x3b, 0x30, 0x3b, 0x38, 0x71, 0x22, 0x31, 0x3b, 0x31})
	w.Write(s.palette)
	for i := from; i < to; i++ {
		w.Write(s.rows[i])
	}
	// string terminator(ST)
	w.Write([]byte{0x1b, 0x5c})
}

func (s *Sixel) Write(w io.Writer) {
	s.write(w, 0, len(s.rows))
}

func (s *Sixel) WriteLeaveUpper(w io.Writer, leaveRows int) {
	s.write(w, leaveRows, len(s.rows))
}

func (s *Sixel) WriteLeaveLower(w io.Writer, leaveRows int) {
	s.write(w, 0, len(s.rows)-leaveRows)
}

func EncodeSixel(img image.Image) *Sixel {
	nc := 255 // (>= 2, 8bit, index 0 is reserved for transparent key color)
	if nc < 2 {
		nc = 255
	}

	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	if width == 0 || height == 0 {
		return nil
	}

	var paletted *image.Paletted

	// fast path for paletted images
	if p, ok := img.(*image.Paletted); ok && len(p.Palette) < int(nc) {
		paletted = p
	} else {
		// make adaptive palette using median cut alogrithm
		q := median.Quantizer(nc - 1)
		paletted = q.Paletted(img)
		draw.Draw(paletted, img.Bounds(), img, image.Point{}, draw.Over)
	}

	pw := bytes.NewBuffer(nil)
	for n, v := range paletted.Palette {
		r, g, b, _ := v.RGBA()
		r = r * 100 / 0xFFFF
		g = g * 100 / 0xFFFF
		b = b * 100 / 0xFFFF
		// DECGCI (#): Graphics Color Introducer
		fmt.Fprintf(pw, "#%d;2;%d;%d;%d", n+1, r, g, b)
	}

	buf := make([]byte, width*nc)
	cset := make([]bool, nc)
	ch0 := specialChNr
	rws := make([][]byte, (height+5)/6)
	w := bytes.NewBuffer(nil)
	for z := 0; z < (height+5)/6; z++ {
		// DECGNL (-): Graphics Next Line
		if z > 0 {
			w.Write([]byte{0x2d})
		}
		for p := 0; p < 6; p++ {
			y := z*6 + p
			for x := 0; x < width; x++ {
				_, _, _, alpha := img.At(x, y).RGBA()
				if alpha != 0 {
					idx := paletted.ColorIndexAt(x, y) + 1
					cset[idx] = false // mark as used
					buf[width*int(idx)+x] |= 1 << uint(p)
				}
			}
		}
		for n := 1; n < nc; n++ {
			if cset[n] {
				continue
			}
			cset[n] = true
			// DECGCR ($): Graphics Carriage Return
			if ch0 == specialChCr {
				w.Write([]byte{0x24})
			}
			// select color (#%d)
			if n >= 100 {
				digit1 := n / 100
				digit2 := (n - digit1*100) / 10
				digit3 := n % 10
				c1 := byte(0x30 + digit1)
				c2 := byte(0x30 + digit2)
				c3 := byte(0x30 + digit3)
				w.Write([]byte{0x23, c1, c2, c3})
			} else if n >= 10 {
				c1 := byte(0x30 + n/10)
				c2 := byte(0x30 + n%10)
				w.Write([]byte{0x23, c1, c2})
			} else {
				w.Write([]byte{0x23, byte(0x30 + n)})
			}
			cnt := 0
			for x := 0; x < width; x++ {
				// make sixel character from 6 pixels
				ch := buf[width*n+x]
				buf[width*n+x] = 0
				if ch0 < 0x40 && ch != ch0 {
					// output sixel character
					s := 63 + ch0
					for ; cnt > 255; cnt -= 255 {
						w.Write([]byte{0x21, 0x32, 0x35, 0x35, s})
					}
					if cnt == 1 {
						w.Write([]byte{s})
					} else if cnt == 2 {
						w.Write([]byte{s, s})
					} else if cnt == 3 {
						w.Write([]byte{s, s, s})
					} else if cnt >= 100 {
						digit1 := cnt / 100
						digit2 := (cnt - digit1*100) / 10
						digit3 := cnt % 10
						c1 := byte(0x30 + digit1)
						c2 := byte(0x30 + digit2)
						c3 := byte(0x30 + digit3)
						// DECGRI (!): - Graphics Repeat Introducer
						w.Write([]byte{0x21, c1, c2, c3, s})
					} else if cnt >= 10 {
						c1 := byte(0x30 + cnt/10)
						c2 := byte(0x30 + cnt%10)
						// DECGRI (!): - Graphics Repeat Introducer
						w.Write([]byte{0x21, c1, c2, s})
					} else if cnt > 0 {
						// DECGRI (!): - Graphics Repeat Introducer
						w.Write([]byte{0x21, byte(0x30 + cnt), s})
					}
					cnt = 0
				}
				ch0 = ch
				cnt++
			}
			if ch0 != 0 {
				// output sixel character
				s := 63 + ch0
				for ; cnt > 255; cnt -= 255 {
					w.Write([]byte{0x21, 0x32, 0x35, 0x35, s})
				}
				if cnt == 1 {
					w.Write([]byte{s})
				} else if cnt == 2 {
					w.Write([]byte{s, s})
				} else if cnt == 3 {
					w.Write([]byte{s, s, s})
				} else if cnt >= 100 {
					digit1 := cnt / 100
					digit2 := (cnt - digit1*100) / 10
					digit3 := cnt % 10
					c1 := byte(0x30 + digit1)
					c2 := byte(0x30 + digit2)
					c3 := byte(0x30 + digit3)
					// DECGRI (!): - Graphics Repeat Introducer
					w.Write([]byte{0x21, c1, c2, c3, s})
				} else if cnt >= 10 {
					c1 := byte(0x30 + cnt/10)
					c2 := byte(0x30 + cnt%10)
					// DECGRI (!): - Graphics Repeat Introducer
					w.Write([]byte{0x21, c1, c2, s})
				} else if cnt > 0 {
					// DECGRI (!): - Graphics Repeat Introducer
					w.Write([]byte{0x21, byte(0x30 + cnt), s})
				}
			}
			ch0 = specialChCr
		}
		rws[z] = make([]byte, w.Len())
		copy(rws[z], w.Bytes())
		w.Reset()
	}

	return &Sixel{
		palette: pw.Bytes(),
		rows:    rws,
	}
}
