package jar

import "strings"

// domainMatch reports whether host matches the cookie domain.
func domainMatch(host, domain string, hostOnly bool) bool {
	if hostOnly {
		return host == domain
	}
	return host == domain || strings.HasSuffix(host, "."+domain)
}

// pathMatch implements RFC 6265 path matching. Query strings are
// excluded by the caller since only the URL path is passed in.
func pathMatch(reqPath, cookiePath string) bool {
	if reqPath == cookiePath {
		return true
	}
	if strings.HasPrefix(reqPath, cookiePath) {
		if strings.HasSuffix(cookiePath, "/") {
			return true
		}
		if len(reqPath) > len(cookiePath) && reqPath[len(cookiePath)] == '/' {
			return true
		}
	}
	return false
}
