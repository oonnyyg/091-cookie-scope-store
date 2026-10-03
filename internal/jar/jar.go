// Package jar implements an offline cookie store with scoped matching
// and JSON-file persistence.
package jar

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Cookie is one stored record.
type Cookie struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Domain     string `json:"domain"` // normalized: lower-case, no leading dot
	Path       string `json:"path"`
	Secure     bool   `json:"secure"`
	HostOnly   bool   `json:"host_only"`
	Persistent bool   `json:"persistent"`
	ExpiresAt  int64  `json:"expires_at"` // absolute seconds, valid when Persistent
	Created    int64  `json:"created"`    // creation order
}

// Input is a cookie as supplied by a receive operation.
type Input struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
	Secure bool   `json:"secure"`
	MaxAge *int64 `json:"max_age"`
}

// Jar holds cookies in creation order.
type Jar struct {
	cookies []*Cookie
	nextID  int64
}

func New() *Jar { return &Jar{} }

func isASCIIAlnum(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

func isASCIIDomain(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

// parseURL validates the URL and returns its lower-case host and path.
func parseURL(raw string) (scheme, host, path string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "", "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	host = u.Hostname()
	if host == "" {
		return "", "", "", errors.New("missing host")
	}
	if net.ParseIP(host) != nil {
		return "", "", "", errors.New("IP hosts are not supported")
	}
	if !isASCIIDomain(host) {
		return "", "", "", errors.New("host must be an ASCII domain")
	}
	host = strings.ToLower(host)
	path = u.EscapedPath()
	if path == "" {
		path = "/"
	}
	return u.Scheme, host, path, nil
}

// defaultPath derives the cookie path from the request path: everything
// before the last slash, or "/" when that is empty.
func defaultPath(reqPath string) string {
	i := strings.LastIndex(reqPath, "/")
	if i <= 0 {
		return "/"
	}
	return reqPath[:i]
}
