package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"math"
	"os/exec"
	"strings"
	"unicode/utf8"

	"git.sr.ht/~ghost08/photon/lib"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/skratchdot/open-golang/open"
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

	imgSixel        *Sixel
	scaledImgBounds image.Rectangle
	underImageRune  rune
	hint            *string
	hints           map[string]string
}

func (a *Article) Draw(ctx Context, s tcell.Screen, sixelBuf *bytes.Buffer) (statusBarText richtext) {
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
		a.HintsParse()
	}
	articleWidthPixels := articleWidth * ctx.XCellPixels()
	x := (ctx.Width - articleWidth) / 2
	contentY := 7

	//header
	drawLinesWordwrap(s, x, 1, articleWidth, 2, a.Title, tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke).Bold(true))
	drawLine(s, x, 4, articleWidth, a.SiteName, tcell.StyleDefault)

	//top image
	if a.TopImage != nil {
		if a.imgSixel == nil {
			imageProcMap.Delete(a)
			imageProc(
				a,
				a.TopImage,
				articleWidthPixels,
				articleWidthPixels,
				func(b image.Rectangle, s *Sixel) {
					a.scaledImgBounds, a.imgSixel = b, s
					redraw(true)
				},
			)
		} else {
			if a.scrollOffset*ctx.YCellPixels() < a.scaledImgBounds.Dy() {
				imageCenterOffset := (articleWidthPixels - a.scaledImgBounds.Dx()) / ctx.XCellPixels() / 2
				setCursorPos(sixelBuf, x+1+imageCenterOffset, contentY)
				leaveRows := int(math.Ceil(float64(a.scrollOffset*ctx.YCellPixels())/6.0)) + 1
				a.imgSixel.WriteLeaveUpper(sixelBuf, leaveRows)
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
						contentY+a.scaledImgBounds.Dy()/ctx.YCellPixels()-a.scrollOffset,
					),
					a.underImageRune,
				)
			}
		}
	} else if a.Article.Article.Image == "" && a.Card.ItemImage != nil {
		imageProcMap.Delete(a)
		imageProc(
			a,
			a.Card.ItemImage,
			articleWidthPixels,
			articleWidthPixels,
			func(b image.Rectangle, s *Sixel) {
				a.scaledImgBounds, a.imgSixel = b, s
				redraw(true)
			},
		)
	}

	//content
	imageYCells := a.scaledImgBounds.Dy() / ctx.YCellPixels()
	for i := max(0, a.scrollOffset-imageYCells); i < len(a.contentLines); i++ {
		line := a.contentLines[i]
		var lineOffset int
		var texts []string
		for _, to := range line {
			if a.hints != nil && a.hint != nil {
				if hint, ok := a.hints[to.Link]; ok && strings.HasPrefix(hint, *a.hint) {
					log.Println(hint, *a.hint)
					s.SetContent(
						x+lineOffset,
						contentY+max(0, imageYCells-a.scrollOffset),
						[]rune(hint)[len(*a.hint)],
						nil,
						tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack),
					)
					lineOffset++
					_, s := utf8.DecodeRuneInString(to.Text)
					to.Text = to.Text[s:]
				}
			}
			lineOffset += drawString(
				s,
				x+lineOffset,
				contentY+max(0, imageYCells-a.scrollOffset),
				to.Text,
				to.Style,
			)
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

func (a *Article) Scroll(d int) {
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

func (a *Article) Clear() {
	a.contentLines = nil
	a.imgSixel = nil
}

func (a *Article) HintStart() {
	var tmp string
	a.hint = &tmp
}

func (a *Article) HintEvent(ev *tcell.EventKey) bool {
	if a == nil || a.hint == nil {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		a.hint = nil
		redraw(false)
	case tcell.KeyRune:
		hint := (*a.hint) + string(ev.Rune())
		a.hint = &hint
		if link := a.HintsFind(); link != "" {
			//found exact hint, openning link and exiting hints mode
			open.Start(link)
			a.hint = nil
		} else if !a.HintsPrefix() {
			//no hints found with set prefix, exiting hints mode
			a.hint = nil
		}
		redraw(false)
	}
	return true
}

func (a *Article) HintsFind() string {
	for link, hint := range a.hints {
		if *a.hint == hint {
			return link
		}
	}
	return ""
}

//returns true if there is a hint with the set prefix
func (a *Article) HintsPrefix() bool {
	for _, hint := range a.hints {
		if strings.HasPrefix(hint, *a.hint) {
			return true
		}
	}
	return false
}

func (a *Article) HintsParse() {
	a.hints = nil
	var (
		chars      = []rune("asdfghjkl")
		minChars   = 1
		shortCount = 0
	)

	var elems []string
	for _, line := range a.contentLines {
		for _, to := range line {
			if to.Link != "" {
				elems = append(elems, to.Link)
			}
		}
	}
	if len(elems) == 0 {
		return
	}
	/*
	   Determine how many digits the link hints will require in the worst
	   case. Usually we do not need all of these digits for every link
	   single hint, so we can show shorter hints for a few of the links.
	*/
	needed := max(minChars, ceilLog(len(elems), len(chars)))

	/*
	   Short hints are the number of hints we can possibly show which are
	   (needed - 1) digits in length.
	*/
	if needed > minChars && needed > 1 {
		totalSpace := int(math.Pow(float64(len(chars)), float64(needed)))
		/*
		 For each 1 short link being added, len(chars) long links are
		 removed, therefore the space removed is len(chars) - 1.
		*/
		shortCount = (totalSpace - len(elems)) // (len(chars) - 1)
	} else {
		shortCount = 0
	}

	longCount := len(elems) - shortCount

	var strings []string

	if needed > 1 {
		for i := 0; i < shortCount; i++ {
			strings = append(strings, numberToHintStr(i, chars, needed-1))
		}
	}

	start := shortCount * len(chars)
	for i := start; i < start+longCount; i++ {
		strings = append(strings, numberToHintStr(i, chars, needed))
	}

	a.hints = make(map[string]string)
	for i, hint := range shuffleHints(strings, len(chars)) {
		if i >= len(elems) {
			break
		}
		a.hints[elems[i]] = hint
	}
}

//Compute max(1, ceil(log(number, base))).  Use only integer arithmetic in order to avoid numerical error.
func ceilLog(number, base int) int {
	if number < 1 || base < 2 {
		panic("math domain error")
	}
	result := 1
	accum := base
	for accum < number {
		result += 1
		accum *= base
	}
	return result
}

/*
	Shuffle the given set of hints so that they're scattered.
	Hints starting with the same character will be spread evenly throughout
	the array.
	Inspired by Vimium.
	Args:
		hints: A list of hint strings.
		length: Length of the available charset.
	Return:
		A list of shuffled hint strings.
*/
func shuffleHints(hints []string, length int) []string {
	buckets := make([][]string, length)
	for i, hint := range hints {
		buckets[i%len(buckets)] = append(buckets[i%len(buckets)], hint)
	}
	var result []string
	for _, bucket := range buckets {
		result = append(result, bucket...)
	}
	return result
}

/*
   Convert a number like "8" into a hint string like "JK".
   This is used to sequentially generate all of the hint text.
   The hint string will be "padded with zeroes" to ensure its length is >=
   digits.
   Inspired by Vimium.
   Args:
       number: The hint number.
       chars: The charset to use.
       digits: The minimum output length.
   Return:
       A hint string.
*/
func numberToHintStr(number int, chars []rune, digits int) string {
	base := len(chars)
	var hintstr []rune
	remainder := 0
	for {
		remainder = number % base
		hintstr = append([]rune{chars[remainder]}, hintstr...)
		number -= remainder
		number = number / base
		if number <= 0 {
			break
		}
	}
	//Pad the hint string we're returning so that it matches digits.
	for i := 0; i < digits-len(hintstr); i++ {
		hintstr = append([]rune{chars[0]}, hintstr...)
	}
	return string(hintstr)
}

type richtext []textobject

type textobject struct {
	Text  string
	Style tcell.Style
	Link  string
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
					to.Link = href
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
				line = append(line, textobject{Text: txt.String(), Style: to.Style, Link: to.Link})
				lines = append(lines, line)
				line = nil
				lineLength = 0
				txt.Reset()
				word.Reset()
				wordLength = 0
				continue
			}
			if c == '\n' || lineLength+wordLength > width {
				line = append(line, textobject{Text: txt.String(), Style: to.Style, Link: to.Link})
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
			line = append(line, textobject{Text: txt.String(), Style: to.Style, Link: to.Link})
			lines = append(lines, line)
			line = richtext{textobject{
				Text:  word.String() + " ",
				Style: to.Style,
				Link:  to.Link,
			}}
			txt.Reset()
			word.Reset()
			lineLength = wordLength + 1
			wordLength = 0
		} else {
			txt.WriteString(word.String())
			txt.WriteRune(' ')
			line = append(line, textobject{Text: txt.String(), Style: to.Style, Link: to.Link})
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
	return string(r)
}
