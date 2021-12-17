package media

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type Extractor struct {
	ExtractorCmd string
	VideoCmd     string
	ImageCmd     string
	TorrentCmd   string
	Client       *http.Client
}

type Media struct {
	e            *Extractor
	OriginalLink string
	Links        []string
	ContentType  string
}

func (e *Extractor) NewMedia(link string) (*Media, error) {
	//if link is a magnet link, or the content type is a torrent file, don't run the extractor
	ct, err := e.getContentType(link)
	if err != nil {
		return nil, fmt.Errorf("media link - getting content-type: %w ", err)
	}
	if ct == "application/x-bittorrent" || ct == "magnet-link" {
		return &Media{e: e, OriginalLink: link, Links: []string{link}, ContentType: ct}, nil
	}
	cmd := strings.Split(
		strings.TrimSpace(strings.ReplaceAll(e.ExtractorCmd, "%", link)),
		" ",
	)
	output, err := exec.Command(cmd[0], cmd[1:]...).Output()
	if err != nil {
		return nil, fmt.Errorf("extracting media link: %w (%s)", err, string(output))
	}
	outputLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var links []string
	for _, line := range outputLines {
		l, err := url.ParseRequestURI(strings.TrimSpace(line))
		if err != nil {
			continue
		}
		links = append(links, l.String())
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("extracting media link: no links extracted")
	}
	contentType, err := e.getContentType(links[0])
	if err != nil {
		return nil, fmt.Errorf("getting media link content-type: %w", err)
	}
	return &Media{e: e, OriginalLink: link, Links: links, ContentType: contentType}, nil
}

func (e *Extractor) getContentType(link string) (string, error) {
	if strings.HasPrefix(link, "magnet:") {
		return "magnet-link", nil
	}
	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		return "", fmt.Errorf("creating HEAD request for content-type detection: %w", err)
	}
	resp, err := e.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending HEAD request for content-type detection: %w", err)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return "", fmt.Errorf("HEAD request doesn't contain content-type")
	}
	return contentType, nil
}

//determineCommand returns videoCmd or imgCmd by the content-type
func (e *Extractor) determineCommand(contentType string) string {
	switch {
	case strings.HasPrefix(contentType, "video/"), contentType == "image/gif", strings.HasSuffix(contentType, "mpegurl"):
		return e.VideoCmd
	case strings.HasPrefix(contentType, "image/"):
		return e.ImageCmd
	case contentType == "application/x-bittorrent", contentType == "magnet-link":
		return e.TorrentCmd
	}
	return ""
}

func (media *Media) Run() {
	command := strings.TrimSpace(media.e.determineCommand(media.ContentType))
	if command == "" {
		log.Println("ERROR: could not determine content-type:", media.ContentType)
		return
	}
	//run command with downloaded torrent file
	if media.ContentType == "application/x-bittorrent" {
		req, err := http.NewRequest("GET", media.Links[0], nil)
		if err != nil {
			log.Printf("ERROR: downloading torrent file - creating http request: %s", err)
			return
		}
		resp, err := media.e.Client.Do(req)
		if err != nil {
			log.Printf("ERROR: downloading torrent file: %s", err)
			return
		}
		defer resp.Body.Close()
		f, err := os.CreateTemp("", "*.torrent")
		if err != nil {
			log.Printf("ERROR: downloading torrent file - creating temp file: %s", err)
			return
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			log.Printf("ERROR: downloading torrent file - writing data to file: %s", err)
			return
		}
		if err := f.Close(); err != nil {
			log.Printf("ERROR: downloading torrent file - closing file: %s", err)
			return
		}
		cmd := strings.Split(
			strings.ReplaceAll(
				command,
				"%",
				f.Name(),
			),
			" ",
		)
		exec.Command(cmd[0], cmd[1:]...).Run()
		return
	}

	//run command with the media link
	if strings.Contains(command, "%") {
		args := media.Links[0]
		if len(media.Links) > 1 {
			args = fmt.Sprintf("%s --audio-file=%s", media.Links[0], media.Links[1])
		}
		cmd := strings.Split(
			strings.ReplaceAll(
				command,
				"%",
				args,
			),
			" ",
		)
		exec.Command(cmd[0], cmd[1:]...).Run()
		return
	}
	//run command with the direct item link
	if strings.Contains(command, "$") {
		cmd := strings.Split(
			strings.ReplaceAll(
				command,
				"$",
				media.OriginalLink,
			),
			" ",
		)
		exec.Command(cmd[0], cmd[1:]...).Run()
		return
	}
	//run command and pipe the media data to it's stdin
	req, err := http.NewRequest("GET", media.Links[0], nil)
	if err != nil {
		log.Println("ERROR: creating GET request for media link:", err)
		return
	}
	resp, err := media.e.Client.Do(req)
	if err != nil {
		log.Println("ERROR: sending GET request for media link:", err)
		return
	}
	cmd := strings.Split(command, " ")
	c := exec.Command(cmd[0], cmd[1:]...)
	stdin, err := c.StdinPipe()
	if err != nil {
		log.Println("ERROR: getting stdin of command:", err)
		return
	}
	go func() {
		defer stdin.Close()
		defer resp.Body.Close()
		io.Copy(stdin, resp.Body)
	}()
	c.Run()
}
