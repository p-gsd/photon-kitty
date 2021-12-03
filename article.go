package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"strings"
	"unicode"

	"git.sr.ht/~ghost08/libphoton"
	"github.com/gdamore/tcell/v2"
	"golang.org/x/net/html"
)

var openedArticle *Article

type Article struct {
	*libphoton.Article
	offset            int
	lastLineDrawn     int
	buffer            []textobject
	topImageSixel     []byte
	scaledImageBounds image.Rectangle
}

func (a *Article) Draw(ctx Context, s tcell.Screen) (buf *bytes.Buffer) {
	s.Clear()
	if a.buffer == nil {
		a.parseArticle(ctx)
	}
	articleWidth := min(72, ctx.Width)
	x := (ctx.Width - articleWidth) / 2
	contentY := 5

	//header
	drawLines(
		s,
		x,
		1,
		articleWidth,
		2,
		a.Title,
		tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke).Bold(true),
	)

	//top image
	if a.TopImage != nil {
		if a.topImageSixel == nil {
			imageProc(
				a,
				a.TopImage,
				articleWidth*int(ctx.XPixel/ctx.Cols),
				articleWidth*int(ctx.XPixel/ctx.Cols),
				func(b image.Rectangle, sd []byte) {
					a.scaledImageBounds, a.topImageSixel = b, sd
					redraw(true)
				},
			)
		} else {
			buf = bytes.NewBuffer(nil)
			fmt.Fprintf(buf, "\033[%d;%dH", contentY, x+1) //set cursor to x, y
			buf.Write(a.topImageSixel)
			contentY += a.scaledImageBounds.Dy()/int(ctx.YPixel/ctx.Rows) + 1
		}
	}

	//content
	for i := -a.offset; i < len(a.buffer); i++ {
		to := a.buffer[i]
		drawLines(
			s,
			x,
			contentY,
			articleWidth,
			ctx.Height,
			to.text,
			to.style,
		)
		a.lastLineDrawn = i
		contentY++
		if contentY > int(ctx.Rows) {
			break
		}
	}
	return
}

func (a *Article) scroll(d int) {
	if a.lastLineDrawn == len(a.buffer)-1 && d < 0 {
		return
	}
	a.offset += d
	if a.offset > 0 {
		a.offset = 0
	}
	if a.offset >= len(a.buffer) {
		a.offset = len(a.buffer) - 1
	}
}

func (a *Article) parseArticle(ctx Context) {
	buf, err := parseArticleContent(a.Node)
	if err != nil {
		log.Println(err)
		return
	}
	var rt richtext
	articleWidth := min(72, ctx.Width)
	var i int
	var txt []rune
	for _, to := range buf {
		for _, c := range to.text {
			if c != '\n' {
				txt = append(txt, c)
			}
			if i == articleWidth || c == '\n' {
				rt = append(rt, textobject{
					text:  string(txt),
					style: to.style,
				})
				i = 0
				txt = nil
				continue
			}
			i++
		}
		rt = append(rt, textobject{
			text:  string(txt),
			style: to.style,
		})
		txt = nil
	}
	a.buffer = rt
}

type richtext []textobject

type textobject struct {
	text  string
	style tcell.Style
}

func parseArticleContent(node *html.Node) (rt richtext, err error) {
	for node := node.FirstChild; node != nil; node = node.NextSibling {
		switch node.Data {
		case "html", "body", "header", "form":
			subrt, err := parseArticleContent(node)
			if err != nil {
				return nil, fmt.Errorf("parsing node <%s>: %w", node.Data, err)
			}
			rt = append(rt, subrt...)
		case "p", "section":
			subrt, err := parseArticleContent(node)
			if err != nil {
				return nil, fmt.Errorf("parsing node <%s>: %w", node.Data, err)
			}
			rt = append(rt, subrt...)
			rt = append(rt, textobject{text: "\n\n", style: tcell.StyleDefault})
		case "span":
			color := tcell.ColorWhite
			for _, attr := range node.Attr {
				if attr.Key == "itemprop" && attr.Val == "description" {
					color = tcell.ColorDarkGray
				}
			}
			subrt, err := parseArticleContent(node)
			if err != nil {
				return nil, fmt.Errorf("parsing node <%s>: %w", node.Data, err)
			}
			subrt = maprt(
				subrt,
				func(to textobject) textobject {
					to.style = tcell.StyleDefault.Foreground(color)
					return to
				},
			)
			rt = append(rt, subrt...)
		case "a":
			subrt, err := parseArticleContent(node)
			if err != nil {
				return nil, fmt.Errorf("parsing node <%s>: %w", node.Data, err)
			}
			var href string
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					href = attr.Val
					break
				}
			}
			subrt = maprt(
				subrt,
				func(to textobject) textobject {
					to.style = to.style.Foreground(tcell.ColorOrangeRed).Bold(true)
					to.text = href
					return to
				},
			)
			rt = append(rt, subrt...)
		case "i", "em", "blockquote", "small":
			subrt, err := parseArticleContent(node)
			if err != nil {
				return nil, fmt.Errorf("parsing node <%s>: %w", node.Data, err)
			}
			subrt = maprt(
				subrt,
				func(to textobject) textobject {
					to.style = to.style.Italic(true)
					return to
				},
			)
			rt = append(rt, subrt...)
		case "strong", "b":
			subrt, err := parseArticleContent(node)
			if err != nil {
				return nil, fmt.Errorf("parsing node <%s>: %w", node.Data, err)
			}
			subrt = maprt(
				subrt,
				func(to textobject) textobject {
					to.style = to.style.Bold(true)
					return to
				},
			)
			rt = append(rt, subrt...)
		case "h1", "h2", "h3":
			subrt, err := parseArticleContent(node)
			if err != nil {
				return nil, fmt.Errorf("parsing node <%s>: %w", node.Data, err)
			}
			subrt = maprt(
				subrt,
				func(to textobject) textobject {
					to.style = to.style.Bold(true)
					return to
				},
			)
			rt = append(rt, subrt...)
			rt = append(rt, textobject{
				style: tcell.StyleDefault.Bold(true),
				text:  "\n\n",
			})
		case "div":
			divrt, err := parseArticleContent(node)
			if err != nil {
				return nil, fmt.Errorf("parsing node <%s>: %w", node.Data, err)
			}
			rt = append(rt, divrt...)
		case "svg", "img", "meta", "head", "hr":
		default:
			if node != nil && node.Type == html.TextNode {
				rt = append(rt, textobject{
					style: tcell.StyleDefault,
					text:  strings.TrimSpace(node.Data),
				})
				log.Println(strings.TrimSpace(node.Data))
				continue
			}
		}
	}
	//add spaces between TextObjects
	if len(rt) > 1 {
		prev := rt[len(rt)-2]
		contentRunes := []rune(prev.text)
		if len(contentRunes) > 0 && !unicode.IsSpace(contentRunes[len(contentRunes)-1]) {
			prev.text += " "
		}
	}
	return rt, nil
}

func maprt(rts []textobject, f func(textobject) textobject) []textobject {
	res := make([]textobject, len(rts))
	for i, rt := range rts {
		res[i] = f(rt)
	}
	return res
}
