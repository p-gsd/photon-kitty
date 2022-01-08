package lib

import (
	"fmt"
	"image"
	"net/http"
	"net/url"
	"time"

	"github.com/cixtor/readability"
)

type Article struct {
	*readability.Article
	Card     *Card
	TopImage image.Image
}

func newArticle(card *Card, client *http.Client) (*Article, error) {
	req, err := http.NewRequest("GET", card.Item.Link, nil)
	if err != nil {
		return nil, err
	}
	uri, err := url.Parse(card.Item.Link)
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
		req.Header.Set("cookie", "__cfduid=d73722d8bb11742b3676371f1c97f19d11517447820")
		req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Safari/537.36")
	}
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("host", uri.Host)
	req.Header.Set("authority", uri.Host)
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("connection", "keep-alive")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("upgrade-insecure-requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if mime := resp.Header.Get("Content-Type"); !isValidContentType(mime) {
		return nil, fmt.Errorf("invalid content-type `%s`", mime)
	}

	a, err := readability.New().Parse(resp.Body, card.Item.Link)
	if err != nil {
		return nil, err
	}
	return &Article{Article: &a, Card: card}, nil
}

func isValidContentType(mime string) bool {
	switch mime {
	case "text/html":
		return true
	case "application/rss+xml":
		return true
	case "text/html;charset=utf-8":
		return true
	case "text/html;charset=UTF-8":
		return true
	case "text/html; charset=utf-8":
		return true
	case "text/html; charset=UTF-8":
		return true
	case "text/html; charset=iso-8859-1":
		return true
	case "text/html; charset=ISO-8859-1":
		return true
	case "text/html; charset=\"utf-8\"":
		return true
	case "text/html; charset=\"UTF-8\"":
		return true
	case "text/html;;charset=UTF-8":
		return true
	case "text/plain; charset=utf-8":
		return true
	case "text/plain; charset=UTF-8":
		return true
	case "application/rss+xml; charset=utf-8":
		return true
	default:
		return false
	}
}
