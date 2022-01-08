package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"math"
	"os/exec"
	"strings"

	"git.sr.ht/~ghost08/photon/lib"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"golang.org/x/net/html"
)

var openedArticle *Article

type ArticleMode int

const (
	ArticleContent ArticleMode = iota
	CardDescription
	CardContent
)

func (as ArticleMode) String() string {
	switch as {
	case ArticleContent:
		return "ARTICLE"
	case CardDescription:
		return "DESCRIPTION"
	case CardContent:
		return "CONTENT"
	}
	return ""
}

func articleModeFromString(s string) ArticleMode {
	switch s {
	case "ARTICLE":
		return ArticleContent
	case "DESCRIPTION":
		return CardDescription
	case "CONTENT":
		return CardContent
	}
	return ArticleContent
}

type Article struct {
	*lib.Article
	scrollOffset int
	lastLine     int
	contentLines []richtext
	Mode         ArticleMode

	topImageSixel     *Sixel
	scaledImageBounds image.Rectangle
	underImageRune    rune
}

func (a *Article) Draw(ctx Context, s tcell.Screen) (sixelBuf *bytes.Buffer, statusBarText richtext) {
	s.Clear()
	articleWidth := min(72, ctx.Width)
	if a.contentLines == nil {
		switch a.Mode {
		case ArticleContent:
			a.contentLines = richtextFromArticle(a.Node, a.TextContent, articleWidth)
		case CardDescription:
			a.contentLines = richtextFromText(renderArticleContent(a.Card.Item.Description), articleWidth)
		case CardContent:
			a.contentLines = richtextFromText(renderArticleContent(a.Card.Item.Content), articleWidth)
		}
	}
	articleWidthPixels := articleWidth * ctx.XCellPixels()
	x := (ctx.Width - articleWidth) / 2
	contentY := 7

	//header
	drawLinesWordwrap(s, x, 1, articleWidth, 2, a.Title, tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke).Bold(true))
	drawLine(s, x, 4, articleWidth, a.SiteName, tcell.StyleDefault)

	//top image
	if a.TopImage != nil {
		if a.topImageSixel == nil {
			imageProcMap.Delete(a)
			imageProc(
				a,
				a.TopImage,
				articleWidthPixels,
				articleWidthPixels,
				func(b image.Rectangle, s *Sixel) {
					a.scaledImageBounds, a.topImageSixel = b, s
					redraw(true)
				},
			)
		} else {
			if a.scrollOffset*ctx.YCellPixels() < a.scaledImageBounds.Dy() {
				sixelBuf = bytes.NewBuffer(nil)
				imageCenterOffset := (articleWidthPixels - a.scaledImageBounds.Dx()) / ctx.XCellPixels() / 2
				setCursorPos(sixelBuf, x+1+imageCenterOffset, contentY)
				leaveRows := int(math.Ceil(float64(a.scrollOffset*ctx.YCellPixels())/6.0)) + 1
				a.topImageSixel.WriteLeaveUpper(sixelBuf, leaveRows)
				if a.underImageRune == '\u2800' {
					a.underImageRune = '\u2007'
				} else {
					a.underImageRune = '\u2800'
				}
				fillArea(
					s,
					image.Rect(
						x,
						contentY-1,
						x+articleWidth,
						contentY+a.scaledImageBounds.Dy()/ctx.YCellPixels()-a.scrollOffset,
					),
					a.underImageRune,
				)
			}
		}
	}

	//content
	imageYCells := a.scaledImageBounds.Dy() / ctx.YCellPixels()
	for i := max(0, a.scrollOffset-imageYCells); i < len(a.contentLines); i++ {
		line := a.contentLines[i]
		var lineOffset int
		var texts []string
		for _, to := range line {
			lineOffset += drawString(s, x+lineOffset, contentY+max(0, imageYCells-a.scrollOffset), to.Text, to.Style)
			texts = append(texts, to.Text)
		}
		a.lastLine = i
		contentY++
		if contentY+max(0, imageYCells-a.scrollOffset) >= int(ctx.Height) {
			break
		}
	}

	//status bar text - article state + scroll percentage
	above := a.scrollOffset
	below := len(a.contentLines) - a.lastLine - 1
	statusBarText = richtext{
		{Text: a.Mode.String(), Style: tcell.StyleDefault.Foreground(tcell.ColorOrangeRed)},
		{Text: "   ", Style: tcell.StyleDefault},
		{Text: scrollPercentage(above, below), Style: tcell.StyleDefault},
	}
	return
}

func (a *Article) scroll(d int) {
	if a.lastLine+d >= len(a.contentLines) {
		a.scrollOffset = len(a.contentLines) - (a.lastLine - a.scrollOffset) - 1
		a.lastLine = len(a.contentLines)
		return
	}
	a.scrollOffset += d
	if a.scrollOffset < 0 {
		a.scrollOffset = 0
	}
}

func (a *Article) ToggleMode() {
	switch a.Mode {
	case ArticleContent:
		a.Mode = CardDescription
	case CardDescription:
		a.Mode = CardContent
	case CardContent:
		a.Mode = ArticleContent
	}
	a.contentLines = nil
	a.scrollOffset = 0
	a.lastLine = 0
}

type richtext []textobject

type textobject struct {
	Text  string
	Style tcell.Style
	link  string
}

func (rt richtext) Len() (lenght int) {
	for _, to := range rt {
		lenght += len(to.Text)
	}
	return
}

func richtextFromText(text string, width int) []richtext {
	return richtextWordWrap(
		richtext{{Text: text, Style: tcell.StyleDefault}},
		width,
	)
}

func richtextFromArticle(node *html.Node, textContent string, width int) []richtext {
	buf, err := parseArticleContent(node)
	if err != nil {
		log.Println(err)
		return nil
	}
	if len(buf) == 0 {
		buf = richtext{
			{
				Text:  textContent,
				Style: tcell.StyleDefault,
			},
		}
	}
	return richtextWordWrap(buf, width)
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

func richtextWordWrap(buf richtext, width int) []richtext {
	//word wrap with textobjects
	var lines []richtext
	var line richtext
	var lineLength, wordLength int
	var txt, word strings.Builder
	for _, to := range buf {
		for _, c := range to.Text {
			if c != '\n' && c != ' ' && wordLength < width {
				word.WriteRune(c)
				wordLength += runewidth.RuneWidth(c)
				continue
			}
			if lineLength+wordLength == width {
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
			if c == '\n' || lineLength+wordLength > width {
				line = append(line, textobject{Text: txt.String(), Style: to.Style})
				lines = append(lines, line)
				line = nil
				txt.Reset()
				if wordLength > 0 {
					txt.WriteString(word.String())
					if wordLength < width {
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
		if lineLength+wordLength > width {
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
	return lines
}

func maprt(rts []textobject, f func(textobject) textobject) []textobject {
	res := make([]textobject, len(rts))
	for i, rt := range rts {
		res[i] = f(rt)
	}
	return res
}

func renderArticleContent(h string) string {
	if CLI.ArticleRenderer == "" {
		return h
	}
	args := strings.Split(CLI.ArticleRenderer, " ")
	if len(args) == 0 {
		return h
	}
	cmd := args[0]
	args = args[1:]
	c := exec.Command(cmd, args...)
	in, err := c.StdinPipe()
	if err != nil {
		return fmt.Sprintf("ERROR: article renderer - opening stdin: %s", err)
	}
	go func() {
		defer in.Close()
		in.Write([]byte(h))
	}()
	r, err := c.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("ERROR: article renderer: %s (%s)", err, string(r))
	}
	log.Println(string(r))
	return string(r)
}
