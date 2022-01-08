package main

import (
	"fmt"
	"image"
	"io"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func scrollPercentage(above, below int) string {
	switch {
	case below <= 0 && above == 0:
		return "All"
	case below <= 0 && above > 0:
		return "Bot"
	case below > 0 && above <= 0:
		return "Top"
	case above > 1000000:
		return fmt.Sprintf("%d%%", above/((above+below)/100))
	default:
		return fmt.Sprintf("%d%%", above*100/(above+below))
	}
}

func fillArea(s tcell.Screen, rect image.Rectangle, r rune) {
	for x := rect.Min.X; x <= rect.Max.X; x++ {
		for y := rect.Min.Y; y <= rect.Max.Y; y++ {
			s.SetContent(x, y, r, nil, tcell.StyleDefault)
		}
	}
}

func setCursorPos(w io.Writer, x, y int) {
	fmt.Fprintf(w, "\033[%d;%dH", y, x)
}

func drawLine(s tcell.Screen, X, Y, maxWidth int, text string, style tcell.Style) (width int) {
	var x int
	for _, c := range text {
		if c == '\n' {
			return
		}
		if x > maxWidth {
			return
		}
		var comb []rune
		w := runewidth.RuneWidth(c)
		width += w
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}

		s.SetContent(x+X, Y, c, comb, style)
		x += w
	}
	return
}

func drawLinesWordwrap(s tcell.Screen, X, Y, maxWidth, maxLines int, text string, style tcell.Style) {
	var x, y int
	var word strings.Builder
	var wordLength int
	for _, c := range text {
		if c != ' ' && c != '\n' {
			word.WriteRune(c)
			wordLength += runewidth.RuneWidth(c)
			continue
		}
		for wordLength > maxWidth && y < maxLines {
			w := drawLine(s, x+X, y+Y, maxWidth-x, word.String(), style)
			wordRest := word.String()[w:]
			word.Reset()
			word.WriteString(wordRest)
			wordLength -= w
			y++
			x = 0
		}
		if y >= maxLines {
			break
		}
		if x+wordLength > maxWidth {
			y++
			x = 0
		}
		if c == '\n' || x+wordLength == maxWidth {
			drawString(s, x+X, y+Y, word.String(), style)
			word.Reset()
			wordLength = 0
			y++
			x = 0
			continue
		}
		if y >= maxLines {
			break
		}
		x += drawString(s, x+X, y+Y, word.String()+" ", style)
		word.Reset()
		wordLength = 0
	}
	if wordLength == 0 {
		return
	}
	if x+wordLength > maxWidth {
		y++
		x = 0
	}
	if y >= maxLines {
		return
	}
	drawString(s, x+X, y+Y, word.String(), style)
}

func drawString(s tcell.Screen, x, y int, text string, style tcell.Style) (width int) {
	for _, c := range text {
		var comb []rune
		w := runewidth.RuneWidth(c)
		width += w
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}

		s.SetContent(x, y, c, comb, style)
		x += w
	}
	return
}
