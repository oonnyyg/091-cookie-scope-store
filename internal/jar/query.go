package jar

import (
	"sort"
	"strings"
)

// Query returns the Cookie request header for the URL at the given time.
// Expired persistent cookies are dropped from the store.
func (j *Jar) Query(rawURL string, now int64) (string, error) {
	scheme, host, reqPath, err := parseURL(rawURL)
	if err != nil {
		return "", err
	}
	secure := scheme == "https"

	var matched []*Cookie
	kept := j.cookies[:0]
	for _, c := range j.cookies {
		if c.Persistent && now >= c.ExpiresAt {
			continue // expired; drop from the store
		}
		kept = append(kept, c)
		if c.Secure && !secure {
			continue
		}
		if !domainMatch(host, c.Domain, c.HostOnly) {
			continue
		}
		if !pathMatch(reqPath, c.Path) {
			continue
		}
		matched = append(matched, c)
	}
	j.cookies = kept

	// Longer paths first; creation order breaks ties (stable sort).
	sort.SliceStable(matched, func(a, b int) bool {
		return len(matched[a].Path) > len(matched[b].Path)
	})

	parts := make([]string, len(matched))
	for i, c := range matched {
		parts[i] = c.Name + "=" + c.Value
	}
	return strings.Join(parts, "; "), nil
}
