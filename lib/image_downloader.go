package lib

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"runtime"
	"sync"

	_ "golang.org/x/image/webp"
)

type ImgDownloader struct {
	client    *http.Client
	receiveCh chan imgDownloadReq
	imgCache  ImageCache
}

type ImageCache interface {
	Load(interface{}) (interface{}, bool)
	Store(interface{}, interface{})
}

type imgDownloadReq struct {
	URL      string
	Callback func(interface{})
}

func newImgDownloader(client *http.Client) *ImgDownloader {
	d := &ImgDownloader{
		client:    client,
		receiveCh: make(chan imgDownloadReq, 1024),
	}
	reqCh := make(chan imgDownloadReq, 1024)
	//receiver
	go func() {
		for req := range d.receiveCh {
			if i, ok := d.imgCache.Load(req.URL); ok {
				if i != nil {
					if req.Callback != nil {
						req.Callback(i)
					}
				}
				continue
			}
			d.imgCache.Store(req.URL, nil)
			reqCh <- req
		}
	}()
	//download workers
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			client := d.client
			if client == nil {
				client = http.DefaultClient
			}
			for req := range reqCh {
				r, err := http.NewRequest("GET", req.URL, nil)
				if err != nil {
					log.Println("ERROR: creating request for image:", err)
					continue
				}
				resp, err := client.Do(r)
				if err != nil {
					log.Println("ERROR: downloading image:", err)
					continue
				}
				defer resp.Body.Close()
				i, _, err := image.Decode(resp.Body)
				if err != nil {
					log.Println("ERROR: decoding image:", err, req.URL)
					continue
				}
				d.imgCache.Store(req.URL, i)
				if req.Callback != nil {
					if img, ok := d.imgCache.Load(req.URL); ok {
						req.Callback(img)
					} else {
						req.Callback(i)
					}
				}
			}
		}()
	}
	return d
}

func (d *ImgDownloader) Download(url string, callback func(interface{})) {
	if d.imgCache == nil {
		d.imgCache = &sync.Map{}
	}
	img, ok := d.imgCache.Load(url)
	if !ok || img == nil {
		d.receiveCh <- imgDownloadReq{
			URL:      url,
			Callback: callback,
		}
	} else {
		if callback != nil {
			callback(img)
		}
	}
}
