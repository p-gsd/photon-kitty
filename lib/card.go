package lib

import (
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"git.sr.ht/~ghost08/photont/lib/events"
	"git.sr.ht/~ghost08/photont/lib/media"
	"github.com/cixtor/readability"
	"github.com/kennygrant/sanitize"
	"github.com/mmcdole/gofeed"
	"github.com/skratchdot/open-golang/open"
)

type Card struct {
	photon    *Photon
	Item      *gofeed.Item
	ItemImage image.Image
	Feed      *gofeed.Feed
	FeedImage image.Image
	Article   *readability.Article
	Media     *media.Media
}

type Cards []*Card

func (cards Cards) Len() int {
	return len(cards)
}

func (cards Cards) Less(i, k int) bool {
	return cards[i].Item.PublishedParsed.After(*cards[k].Item.PublishedParsed)
}

func (cards Cards) Swap(i, k int) {
	cards[i], cards[k] = cards[k], cards[i]
}

func (card *Card) SaveImage() func(image.Image) {
	return func(img image.Image) {
		card.ItemImage = img
		/*
			TODO check if this is needed
			for _, vi := range onScreenCards {
				if vi == card {
					card.photon.cb.Redraw()
					break
				}
			}
		*/
		card.photon.cb.Redraw()
	}
}

func (card *Card) OpenArticle() {
	if card == nil {
		return
	}
	if card.Article == nil {
		article, err := newArticle(card.Item.Link, card.photon.httpClient)
		if err != nil {
			log.Println("ERROR: scraping link:", err)
			return
		}
		card.Article = article
		if len(card.Article.TextContent) < len(card.Item.Description) {
			card.Article.TextContent = card.Item.Description
		}
	}
	card.photon.OpenedArticle = &Article{Article: card.Article}
	card.photon.OpenedArticle.Link = card.Item.Link
	card.photon.cb.ArticleChanged(card.photon.OpenedArticle)
	if card.photon.OpenedArticle.Image != "" {
		card.photon.ImgDownloader.Download(
			card.photon.OpenedArticle.Image,
			func(img image.Image) {
				if card.photon.OpenedArticle == nil {
					return
				}
				card.photon.OpenedArticle.TopImage = img
				card.photon.cb.Redraw()
			},
		)
	}
	card.photon.cb.Redraw()
}

func (card *Card) GetMedia() (*media.Media, error) {
	if card == nil {
		return nil, nil
	}
	if card.Media == nil || len(card.Media.Links) == 0 {
		m, err := card.photon.mediaExtractor.NewMedia(card.Item.Link)
		if err != nil {
			return nil, err
		}
		card.Media = m
	}
	return card.Media, nil
}

func (card *Card) RunMedia() {
	if card == nil {
		return
	}
	events.Emit(&events.RunMediaStart{Link: card.Item.Link})
	card.photon.cb.Redraw()
	go func() {
		defer func() {
			events.Emit(&events.RunMediaEnd{Link: card.Item.Link})
			card.photon.cb.Redraw()
		}()
		if _, err := card.GetMedia(); err != nil {
			log.Println("ERROR: extracting media link:", err)
			return
		}
		card.Media.Run()
	}()
}

func (card *Card) DownloadMedia() {
	if card == nil {
		return
	}
	go func() {
		log.Println("INFO: downloading media for:", card.Item.Link)
		m, err := card.GetMedia()
		if err != nil {
			log.Println("ERROR: extracting media link:", err)
			return
		}
		if err := card.downloadLinks(card.Item.Title, m.Links); err != nil {
			log.Println("ERROR: downloading media:", err)
			return
		}
	}()
}

func (card *Card) DownloadLink() {
	if card == nil {
		return
	}
	go func() {
		if err := card.downloadLinks(card.Item.Title, []string{card.Item.Link}); err != nil {
			log.Println("ERROR: downloading link:", err)
			return
		}
	}()
}

func (card *Card) DownloadImage() {
	if card == nil || card.Item == nil || card.Item.Image == nil {
		return
	}
	go func() {
		if err := card.downloadLinks(card.Item.Title, []string{card.Item.Image.URL}); err != nil {
			log.Println("ERROR: downloading image:", err)
			return
		}
	}()
}

func (card *Card) downloadLinks(name string, links []string) error {
	//get download path
	downloadPath := card.photon.downloadPath
	if strings.Contains(downloadPath, "$HOME") {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		downloadPath = strings.ReplaceAll(downloadPath, "$HOME", usr.HomeDir)
	}
	//create download path
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		if err := os.MkdirAll(downloadPath, 0x755); err != nil {
			return err
		}
	}
	for _, link := range links {
		//get response
		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			return err
		}
		client := card.photon.httpClient
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var bodyReader io.Reader = resp.Body
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType, bodyReader, err = recycleReader(resp.Body)
		}
		exts := extensionByType(contentType)
		if len(exts) > 0 {
			name += "." + exts[0]
		}
		log.Println(contentType, name)
		//write data to file
		f, err := os.Create(filepath.Join(downloadPath, sanitize.Name(name)))
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, bodyReader); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		log.Printf("INFO: downloaded link %s to %s", link, filepath.Join(downloadPath, name))
	}
	return nil
}

func (card *Card) OpenBrowser() {
	if card == nil {
		return
	}
	open.Start(card.Item.Link)
}
