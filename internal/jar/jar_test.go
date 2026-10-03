package jar

import (
	"testing"
)

func i64(v int64) *int64 { return &v }

func mustReceive(t *testing.T, j *Jar, url string, now int64, in Input) {
	t.Helper()
	if err := j.Receive(url, now, in); err != nil {
		t.Fatalf("Receive(%s, %+v): %v", url, in, err)
	}
}

func mustQuery(t *testing.T, j *Jar, url string, now int64) string {
	t.Helper()
	h, err := j.Query(url, now)
	if err != nil {
		t.Fatalf("Query(%s): %v", url, err)
	}
	return h
}

func TestBasicHostOnlyCookie(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "a", Value: "1"})
	if got := mustQuery(t, j, "http://example.com/", 0); got != "a=1" {
		t.Fatalf("got %q", got)
	}
	// Host-only: subdomains must not match.
	if got := mustQuery(t, j, "http://sub.example.com/", 0); got != "" {
		t.Fatalf("host-only leaked to subdomain: %q", got)
	}
}

func TestDomainCookieMatchesSubdomains(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://sub.example.com/", 0,
		Input{Name: "a", Value: "1", Domain: ".EXAMPLE.com"})
	for _, u := range []string{
		"http://example.com/",
		"http://sub.example.com/",
		"http://deep.sub.example.com/",
	} {
		if got := mustQuery(t, j, u, 0); got != "a=1" {
			t.Fatalf("%s: got %q", u, got)
		}
	}
	if got := mustQuery(t, j, "http://notexample.com/", 0); got != "" {
		t.Fatalf("suffix confusion: %q", got)
	}
	if got := mustQuery(t, j, "http://badexample.com.evil.com/", 0); got != "" {
		t.Fatalf("cross-domain leak: %q", got)
	}
}

func TestDomainRejectedWhenNotSuffixOfOrigin(t *testing.T) {
	j := New()
	err := j.Receive("http://example.com/", 0,
		Input{Name: "a", Value: "1", Domain: "other.com"})
	if err == nil {
		t.Fatal("expected rejection")
	}
	err = j.Receive("http://example.com/", 0,
		Input{Name: "a", Value: "1", Domain: "sub.example.com"})
	if err == nil {
		t.Fatal("expected rejection of child domain")
	}
}

func TestPortIgnored(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com:8080/", 0, Input{Name: "a", Value: "1"})
	if got := mustQuery(t, j, "http://example.com:9090/", 0); got != "a=1" {
		t.Fatalf("port should not matter: %q", got)
	}
}
