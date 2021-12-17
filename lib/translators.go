package lib

import (
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/atom"
	"github.com/mmcdole/gofeed/rss"
	"golang.org/x/net/html"
)

func getExt(f func() string) (val string, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			val = ""
			ok = false
		}
	}()
	return f(), true
}

//find img html element and extract raw text in description and/or content
func scrapContent(item *gofeed.Item) {
	var simpleContent string
	z := html.NewTokenizer(
		strings.NewReader(
			html.UnescapeString(item.Description + "\n" + item.Content),
		),
	)
	previousStartTokenTest := z.Token()
loopDomTest:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break loopDomTest // End of the document,  done
		case html.SelfClosingTagToken:
			if item.Image != nil {
				continue
			}
			t := z.Token()
			if t.Data != "img" {
				continue
			}
			for _, attr := range t.Attr {
				if attr.Key == "src" {
					item.Image = &gofeed.Image{URL: attr.Val}
				}
			}
		case html.StartTagToken:
			previousStartTokenTest = z.Token()
		case html.TextToken:
			switch previousStartTokenTest.Data {
			case "script", "style":
				continue
			}
			TxtContent := strings.TrimSpace(html.UnescapeString(string(z.Text())))
			if len(TxtContent) > 0 {
				simpleContent += TxtContent + "\n"
			}
		}
	}
	item.Description = simpleContent
}

func findImage(item *gofeed.Item) {
	for _, e := range item.Enclosures {
		if strings.HasPrefix(e.Type, "image/") {
			item.Image = &gofeed.Image{URL: item.Enclosures[0].URL}
			return
		}
	}
	if url, ok := getExt(func() string { return item.Extensions["media"]["group"][0].Children["thumbnail"][0].Attrs["url"] }); ok {
		item.Image = &gofeed.Image{URL: url}
		return
	}
	if url, ok := getExt(func() string { return item.Extensions["media"]["preview"][0].Attrs["url"] }); ok {
		item.Image = &gofeed.Image{URL: url}
		return
	}
	if url, ok := item.Custom["thumb_large"]; ok {
		item.Image = &gofeed.Image{URL: url}
		return
	}
	if url, ok := item.Custom["thumb"]; ok {
		item.Image = &gofeed.Image{URL: url}
		return
	}
}

type customAtomTranslator struct {
	defaultTranslator *gofeed.DefaultAtomTranslator
}

func newCustomAtomTranslator() *customAtomTranslator {
	t := &customAtomTranslator{}
	t.defaultTranslator = &gofeed.DefaultAtomTranslator{}
	return t
}

func (ct *customAtomTranslator) Translate(feed interface{}) (*gofeed.Feed, error) {
	atom, found := feed.(*atom.Feed)
	if !found {
		return nil, fmt.Errorf("Feed did not match expected type of *atom.Feed")
	}

	f, err := ct.defaultTranslator.Translate(atom)
	if err != nil {
		return nil, err
	}

	if atom.Icon != "" && f.Image == nil {
		f.Image = &gofeed.Image{URL: strings.TrimSuffix(atom.Icon, "/")}
	}

	for _, i := range f.Items {
		if i.Image == nil || i.Image.URL == "" {
			findImage(i)
		}
		if val, ok := getExt(func() string { return i.Extensions["media"]["group"][0].Children["description"][0].Value }); ok {
			i.Description = val
		}
		scrapContent(i)
		if i.PublishedParsed == nil {
			if i.UpdatedParsed == nil {
				pubdate, ok := i.Custom["pubdate"]
				if !ok {
					i.PublishedParsed = &time.Time{}
					continue
				}
				pd, err := time.Parse(time.RFC3339, pubdate)
				if err != nil {
					return nil, fmt.Errorf("parsing pubdate: %w", err)
				}
				i.PublishedParsed = &pd
			} else {
				i.PublishedParsed = i.UpdatedParsed
			}
		}
	}
	return f, nil
}

type customRSSTranslator struct {
	defaultTranslator *gofeed.DefaultRSSTranslator
}

func newCustomRSSTranslator() *customRSSTranslator {
	t := &customRSSTranslator{}
	t.defaultTranslator = &gofeed.DefaultRSSTranslator{}
	return t
}

func (ct *customRSSTranslator) Translate(feed interface{}) (*gofeed.Feed, error) {
	rss, found := feed.(*rss.Feed)
	if !found {
		return nil, fmt.Errorf("Feed did not match expected type of *rss.Feed")
	}

	f, err := ct.defaultTranslator.Translate(rss)
	if err != nil {
		return nil, err
	}

	for _, i := range f.Items {
		if i.Image == nil || i.Image.URL == "" {
			findImage(i)
		}
		scrapContent(i)
		if i.PublishedParsed == nil {
			if i.UpdatedParsed == nil {
				pubdate, ok := i.Custom["pubdate"]
				if !ok {
					i.PublishedParsed = &time.Time{}
					continue
				}
				pd, err := time.Parse(time.RFC3339, pubdate)
				if err != nil {
					return nil, fmt.Errorf("parsing pubdate: %w", err)
				}
				i.PublishedParsed = &pd
			} else {
				i.PublishedParsed = i.UpdatedParsed
			}
		}
	}
	return f, nil
}
