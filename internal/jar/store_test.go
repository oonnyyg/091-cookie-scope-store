package jar

import (
	"path/filepath"
	"testing"
)

func TestSecureCookieHTTPSOnly(t *testing.T) {
	j := New()
	mustReceive(t, j, "https://example.com/", 0,
		Input{Name: "s", Value: "1", Secure: true})
	if got := mustQuery(t, j, "http://example.com/", 0); got != "" {
		t.Fatalf("secure cookie over http: %q", got)
	}
	if got := mustQuery(t, j, "https://example.com/", 0); got != "s=1" {
		t.Fatalf("secure cookie over https: %q", got)
	}
}

func TestMaxAgeExpiry(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/", 10,
		Input{Name: "p", Value: "1", MaxAge: i64(100)})
	if got := mustQuery(t, j, "http://example.com/", 109); got != "p=1" {
		t.Fatalf("before expiry: %q", got)
	}
	if got := mustQuery(t, j, "http://example.com/", 110); got != "" {
		t.Fatalf("at expiry must be gone: %q", got)
	}
}

func TestMaxAgeNonPositiveDeletes(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "a", Value: "1"})
	mustReceive(t, j, "http://example.com/", 1,
		Input{Name: "a", Value: "", MaxAge: i64(0)})
	if got := mustQuery(t, j, "http://example.com/", 2); got != "" {
		t.Fatalf("delete failed: %q", got)
	}
}

func TestValidationRejectsBadInput(t *testing.T) {
	j := New()
	cases := []struct {
		url string
		in  Input
	}{
		{"ftp://example.com/", Input{Name: "a", Value: "1"}},
		{"http://127.0.0.1/", Input{Name: "a", Value: "1"}},
		{"http://example.com/", Input{Name: "", Value: "1"}},
		{"http://example.com/", Input{Name: "a-b", Value: "1"}},
		{"http://example.com/", Input{Name: "a", Value: "x y"}},
		{"http://example.com/", Input{Name: "a", Value: "1", Domain: "192.168.0.1"}},
	}
	for _, c := range cases {
		if err := j.Receive(c.url, 0, c.in); err == nil {
			t.Fatalf("expected error for %s %+v", c.url, c.in)
		}
	}
	// Errors must not change the store.
	if got := mustQuery(t, j, "http://example.com/", 0); got != "" {
		t.Fatalf("store changed by failed ops: %q", got)
	}
}

func TestPersistenceAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenStore(filepath.Join(dir))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CheckTime(10); err != nil {
		t.Fatal(err)
	}
	// Persistent cookie survives; session cookie does not.
	if err := s.Receive("http://example.com/", 10,
		Input{Name: "p", Value: "1", MaxAge: i64(1000)}); err != nil {
		t.Fatal(err)
	}
	if err := s.Receive("http://example.com/", 10,
		Input{Name: "sess", Value: "2"}); err != nil {
		t.Fatal(err)
	}

	s2, err := OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s2.CheckTime(20); err != nil {
		t.Fatal(err)
	}
	got, err := s2.Query("http://example.com/", 20)
	if err != nil {
		t.Fatal(err)
	}
	if got != "p=1" {
		t.Fatalf("after restart: %q", got)
	}
}

func TestTimeMonotonicAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CheckTime(50); err != nil {
		t.Fatal(err)
	}
	if err := s.Receive("http://example.com/", 50,
		Input{Name: "p", Value: "1", MaxAge: i64(1000)}); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckTime(49); err == nil {
		t.Fatal("expected backwards-time rejection in session")
	}

	s2, err := OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s2.CheckTime(49); err == nil {
		t.Fatal("expected backwards-time rejection after restart")
	}
	if err := s2.CheckTime(50); err != nil {
		t.Fatal(err)
	}
	if err := s2.CheckTime(-1); err == nil {
		t.Fatal("expected negative-time rejection")
	}
}
