package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"strings"

	"git.sr.ht/~ghost08/libphoton"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"golang.org/x/net/html"
)

var openedArticle *Article

type Article struct {
	*libphoton.Article
	firstLine         int
	lastLine          int
	contentLines      []richtext
	topImageSixel     *Sixel
	scaledImageBounds image.Rectangle
}

func (a *Article) Draw(ctx Context, s tcell.Screen) (buf *bytes.Buffer) {
	s.Clear()
	if a.contentLines == nil {
		a.parseArticle(ctx)
	}
	articleWidth := min(72, ctx.Width)
	x := (ctx.Width - articleWidth) / 2
	contentY := 5

	//header
	drawLinesWordwrap(
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
				func(b image.Rectangle, s *Sixel) {
					a.scaledImageBounds, a.topImageSixel = b, s
					redraw(true)
				},
			)
		} else {
			buf = bytes.NewBuffer(nil)
			fmt.Fprintf(buf, "\033[%d;%dH", contentY, x+1) //set cursor to x, y
			a.topImageSixel.Write(buf)
			contentY += a.scaledImageBounds.Dy()/int(ctx.YPixel/ctx.Rows) + 1
		}
	}

	//content
	for i := a.firstLine; i < len(a.contentLines); i++ {
		line := a.contentLines[i]
		var offset int
		var texts []string
		for _, to := range line {
			offset += drawString(
				s,
				x+offset,
				contentY,
				to.Text,
				to.Style,
			)
			texts = append(texts, to.Text)
		}
		a.lastLine = i
		contentY++
		if contentY > int(ctx.Rows) {
			break
		}
	}
	return
}

func (a *Article) scroll(d int) {
	if a.lastLine == len(a.contentLines)-1 && d > 0 {
		return
	}
	a.firstLine += d
	if a.firstLine < 0 {
		a.firstLine = 0
	}
	if a.firstLine >= len(a.contentLines) {
		a.firstLine = len(a.contentLines) - 1
	}
	if a.lastLine+d >= len(a.contentLines) {
		a.firstLine -= a.lastLine + d - len(a.contentLines)
		a.lastLine = len(a.contentLines)
	}
}

func (a *Article) parseArticle(ctx Context) {
	buf, err := parseArticleContent(a.Node)
	if err != nil {
		log.Println(err)
		return
	}
	if len(buf) == 0 {
		buf = richtext{
			{
				Text:  a.TextContent,
				Style: tcell.StyleDefault,
			},
		}
	}

	//word wrap with textobjects
	articleWidth := min(72, ctx.Width)
	var lines []richtext
	var line richtext
	var lineLength, wordLength int
	var txt, word strings.Builder
	for _, to := range buf {
		for _, c := range to.Text {
			if c != '\n' && c != ' ' && wordLength < articleWidth {
				word.WriteRune(c)
				wordLength += runewidth.RuneWidth(c)
				continue
			}
			if lineLength+wordLength == articleWidth {
				if wordLength > 0 {
					txt.WriteString(word.String())
				}
				line = append(line, textobject{Text: txt.String(), Style: to.Style})
				lines = append(lines, line)
				line = nil
				lineLength = 0
				txt.Reset()
				word.Reset()
				wordLength = 0
				continue
			}
			if c == '\n' || lineLength+wordLength > articleWidth {
				line = append(line, textobject{Text: txt.String(), Style: to.Style})
				lines = append(lines, line)
				line = nil
				txt.Reset()
				if wordLength > 0 {
					txt.WriteString(word.String())
					if wordLength < articleWidth {
						txt.WriteRune(' ')
					}
					lineLength = wordLength + 1
					word.Reset()
					wordLength = 0
					continue
				}
				lineLength = 0
				continue
			}
			if wordLength > 0 {
				txt.WriteString(word.String())
				txt.WriteRune(' ')
				lineLength += wordLength + 1
				word.Reset()
				wordLength = 0
			}
		}
		if wordLength == 0 {
			continue
		}
		if lineLength+wordLength > articleWidth {
			line = append(line, textobject{Text: txt.String(), Style: to.Style})
			lines = append(lines, line)
			line = richtext{textobject{
				Text:  word.String() + " ",
				Style: to.Style,
			}}
			txt.Reset()
			word.Reset()
			lineLength = wordLength + 1
			wordLength = 0
		} else {
			txt.WriteString(word.String())
			txt.WriteRune(' ')
			line = append(line, textobject{Text: txt.String(), Style: to.Style})
			txt.Reset()
			word.Reset()
			lineLength += wordLength + 1
			wordLength = 0
		}
	}
	a.contentLines = lines
}

type richtext []textobject

type textobject struct {
	Text  string
	Style tcell.Style
	link  string
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
			rt = append(rt, textobject{Text: "\n\n", Style: tcell.StyleDefault})
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
					to.Style = tcell.StyleDefault.Foreground(color)
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
					to.Style = to.Style.Foreground(tcell.ColorOrangeRed).Bold(true)
					to.link = href
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
					to.Style = to.Style.Italic(true)
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
					to.Style = to.Style.Bold(true)
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
					to.Style = to.Style.Bold(true)
					return to
				},
			)
			rt = append(rt, subrt...)
			rt = append(rt, textobject{
				Style: tcell.StyleDefault.Bold(true),
				Text:  "\n\n",
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
					Style: tcell.StyleDefault,
					Text:  strings.TrimSpace(node.Data),
				})
				continue
			}
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
