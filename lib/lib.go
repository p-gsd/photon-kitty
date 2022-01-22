package lib

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"git.sr.ht/~ghost08/photon/lib/events"
	"git.sr.ht/~ghost08/photon/lib/inputs"
	"git.sr.ht/~ghost08/photon/lib/keybindings"
	"git.sr.ht/~ghost08/photon/lib/media"
	"git.sr.ht/~ghost08/photon/lib/states"
	"github.com/mmcdole/gofeed"
	lua "github.com/yuin/gopher-lua"
)

type Photon struct {
	feedInputs     *inputs.Inputs
	ImgDownloader  *ImgDownloader
	mediaExtractor *media.Extractor
	httpClient     *http.Client
	KeyBindings    *keybindings.Registry
	downloadPath   string
	cb             Callbacks
	luaState       *lua.LState

	Cards               Cards
	VisibleCards        Cards
	SelectedCard        *Card
	SelectedCardPos     image.Point
	searchQuery         string
	OpenedArticle       *Article
	status              string
	statusTimeoutCancel chan struct{}
}

type Callbacks interface {
	Redraw()
	State() states.Enum
	ArticleChanged(*Article)
}

func New(cb Callbacks, paths []string, options ...Option) (*Photon, error) {
	p := &Photon{
		KeyBindings: keybindings.New(cb.State),
		cb:          cb,
	}
	feedInputs, err := p.loadFeeds(paths)
	if err != nil {
		return nil, err
	}
	if feedInputs.Len() == 0 {
		return nil, fmt.Errorf("no feeds")
	}
	p.feedInputs = feedInputs
	p.mediaExtractor = &media.Extractor{Client: p.httpClient}
	p.ImgDownloader = newImgDownloader(p.httpClient)
	for _, o := range options {
		o(p)
	}
	if p.httpClient == nil {
		p.httpClient = http.DefaultClient
	}
	p.mediaExtractor.Client = p.httpClient
	p.ImgDownloader.client = p.httpClient
	if err = p.loadPlugins(); err != nil {
		log.Fatal("ERROR:", err)
	}
	events.Emit(&events.Init{})
	return p, nil
}

type Option func(*Photon)

func WithHTTPClient(c *http.Client) Option {
	return func(p *Photon) {
		p.httpClient = c
	}
}

func WithMediaExtractor(extractor string) Option {
	return func(p *Photon) {
		p.mediaExtractor.ExtractorCmd = extractor
	}
}

func WithMediaVideoCmd(videoCmd string) Option {
	return func(p *Photon) {
		p.mediaExtractor.VideoCmd = videoCmd
	}
}

func WithMediaImageCmd(imageCmd string) Option {
	return func(p *Photon) {
		p.mediaExtractor.ImageCmd = imageCmd
	}
}

func WithMediaTorrentCmd(torrentCmd string) Option {
	return func(p *Photon) {
		p.mediaExtractor.TorrentCmd = torrentCmd
	}
}

func WithDownloadPath(downloadPath string) Option {
	return func(p *Photon) {
		p.downloadPath = downloadPath
	}
}

func (p *Photon) loadFeeds(paths []string) (*inputs.Inputs, error) {
	var ret []string
	for _, path := range paths {
		switch {
		case path == "-":
			if len(paths) > 1 {
				log.Fatal("ERROR: cannot parse from args and stdin")
			}
			ret = append(ret, "-")
		case strings.HasPrefix(path, "http://"), strings.HasPrefix(path, "https://"), strings.HasPrefix(path, "cmd://"):
			ret = append(ret, path)
		default:
			f, err := os.Open(path)
			if err != nil {
				log.Fatal("ERROR: opening file:", err)
			}
			defer f.Close()
			feeds, err := inputs.Parse(f)
			if err != nil {
				log.Fatal("ERROR: parsing file:", err)
			}
			ret = append(ret, feeds...)
		}
	}
	return (*inputs.Inputs)(&ret), nil
}

func (p *Photon) SearchQuery(q string) {
	p.searchQuery = q
	p.filterCards()
}

func (p *Photon) DownloadFeeds() {
	p.Cards = nil
	feeds := make(chan *gofeed.Feed)
	for _, feedURL := range *p.feedInputs {
		feedURL := feedURL
		go func() {
			fp := gofeed.NewParser()
			fp.Client = p.httpClient
			fp.AtomTranslator = newCustomAtomTranslator()
			fp.RSSTranslator = newCustomRSSTranslator()
			var err error
			var f *gofeed.Feed
			switch {
			case feedURL == "-":
				f, err = fp.Parse(os.Stdin)
			case strings.HasPrefix(feedURL, "cmd://"):
				var command []string
				for _, c := range strings.Split(feedURL[6:], " ") {
					c = strings.TrimSpace(c)
					if c != "" {
						command = append(command, c)
					}
				}
				cmd := exec.Command(command[0], command[1:]...)
				var stdout bytes.Buffer
				cmd.Stdout = &stdout
				if err := cmd.Run(); err != nil {
					log.Printf("ERROR: running command (%s): %s", feedURL, err)
					feeds <- nil
					return
				}
				f, err = fp.Parse(&stdout)
			case strings.HasPrefix(feedURL, "http://"), strings.HasPrefix(feedURL, "https://"):
				f, err = fp.ParseURL(feedURL)
			default:
				log.Fatalf("ERROR: not supported feed: %s", feedURL)
			}
			if err != nil {
				log.Printf("ERROR: downloading feed (%s): %s", feedURL, err)
				feeds <- nil
				return
			}
			if f.Image != nil && f.Image.URL != "" {
				p.ImgDownloader.Download(f.Image.URL, nil)
			}
			feeds <- f
		}()
	}
	var (
		feedsGot     int
		spinnerArray = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		spinnerIndex int
	)
	for {
		var f *gofeed.Feed
		select {
		case f = <-feeds:
			feedsGot++
		case <-time.Tick(time.Millisecond * 150):
			spinnerIndex = (spinnerIndex + 1) % len(spinnerArray)
		}
		p.setStatus(
			fmt.Sprintf(
				"Downloading feeds %d/%d %c",
				feedsGot,
				p.feedInputs.Len(),
				spinnerArray[spinnerIndex],
			),
		)
		if f == nil {
			if feedsGot == p.feedInputs.Len() {
				break
			}
			continue
		}
		newCards := make(Cards, len(f.Items))
		for i, item := range f.Items {
			newCards[i] = &Card{
				photon:     p,
				Item:       item,
				Feed:       f,
				Foreground: -1,
				Background: -1,
			}
		}
		p.Cards = append(p.Cards, newCards...)
		if feedsGot == p.feedInputs.Len() {
			break
		}
		f = nil
	}
	p.setStatus("")
	sort.Sort(p.Cards)
	p.filterCards()
	if len(p.VisibleCards) > 0 {
		p.SelectedCard = p.VisibleCards[0]
	}
	events.Emit(&events.FeedsDownloaded{})
}

func (p *Photon) filterCards() {
	query := strings.ToLower(strings.TrimPrefix(p.searchQuery, "/"))
	if query == "" {
		p.VisibleCards = p.Cards
		return
	}
	p.VisibleCards = nil
	for _, card := range p.Cards {
		if strings.Contains(strings.ToLower(card.Item.Title), query) ||
			strings.Contains(strings.ToLower(card.Item.Description), query) ||
			strings.Contains(strings.ToLower(card.Feed.Title), query) ||
			card.Item.Author != nil && strings.Contains(strings.ToLower(card.Item.Author.Name), query) {
			p.VisibleCards = append(p.VisibleCards, card)
		}
	}
}

func (p *Photon) GetStatus() string {
	return p.status
}

func (p *Photon) setStatus(text string) {
	if p.statusTimeoutCancel != nil {
		p.statusTimeoutCancel <- struct{}{}
		p.statusTimeoutCancel = nil
	}
	p.status = text
	p.cb.Redraw()
}

func (p *Photon) StatusWithTimeout(text string, d time.Duration) {
	p.setStatus(text)
	p.statusTimeoutCancel = make(chan struct{})
	go func() {
		select {
		case <-p.statusTimeoutCancel:
			return
		case <-time.After(d):
			p.setStatus("")
		}
	}()
}
