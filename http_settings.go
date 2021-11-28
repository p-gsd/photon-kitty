package main

import (
	"crypto/tls"
	"net/http"
)

type HTTPSettings struct {
	Cookie    map[string]string `optional:"" short:"b" help:"sets the cookie of all outgoing http requests"`
	Header    map[string]string `optional:"" short:"H" help:"sets the headers of all outgoing http requests"`
	UserAgent string            `optional:"" short:"A" help:"sets the User-Agent header of all outgoing http request" default:"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Safari/537.36"`
	Insecure  bool              `optional:"" short:"k" default:"false" help:"(TLS) By default, every SSL connection photon makes is verified to be secure. This option allows photon to proceed and operate even for server connections otherwise considered insecure"`
}

func (s *HTTPSettings) Client() *http.Client {
	client := &http.Client{
		Transport: &photonTransport{
			T: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: s.Insecure},
			},
			s: s,
		},
	}
	return client
}

type photonTransport struct {
	T http.RoundTripper
	s *HTTPSettings
}

func (rt *photonTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.s.Cookie != nil {
		for k, v := range rt.s.Cookie {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}
	if rt.s.Header != nil {
		for k, v := range rt.s.Header {
			req.Header.Set(k, v)
		}
	}
	req.Header.Set("User-Agent", rt.s.UserAgent)
	return rt.T.RoundTrip(req)
}
