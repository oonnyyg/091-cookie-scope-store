package jar

import "testing"

func TestDefaultPath(t *testing.T) {
	j := New()
	// Request path /a/b/c -> default path /a/b.
	mustReceive(t, j, "http://example.com/a/b/c", 0, Input{Name: "a", Value: "1"})
	if got := mustQuery(t, j, "http://example.com/a/b", 0); got != "a=1" {
		t.Fatalf("exact default path: %q", got)
	}
	if got := mustQuery(t, j, "http://example.com/a/b/c", 0); got != "a=1" {
		t.Fatalf("prefix with slash: %q", got)
	}
	if got := mustQuery(t, j, "http://example.com/a/bc", 0); got != "" {
		t.Fatalf("prefix without slash must not match: %q", got)
	}
	if got := mustQuery(t, j, "http://example.com/a", 0); got != "" {
		t.Fatalf("parent must not match: %q", got)
	}
}

func TestDefaultPathRoot(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/x", 0, Input{Name: "a", Value: "1"})
	if got := mustQuery(t, j, "http://example.com/anywhere/deep", 0); got != "a=1" {
		t.Fatalf("root default path: %q", got)
	}
}

func TestExplicitPathValidation(t *testing.T) {
	j := New()
	if err := j.Receive("http://example.com/", 0,
		Input{Name: "a", Value: "1", Path: "noslash"}); err == nil {
		t.Fatal("expected rejection of path without leading slash")
	}
}

func TestPathTrailingSlash(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/", 0,
		Input{Name: "a", Value: "1", Path: "/dir/"})
	if got := mustQuery(t, j, "http://example.com/dir/file", 0); got != "a=1" {
		t.Fatalf("trailing-slash prefix: %q", got)
	}
	if got := mustQuery(t, j, "http://example.com/dir", 0); got != "" {
		t.Fatalf("should not match without trailing slash: %q", got)
	}
}

func TestQueryStringIgnored(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/p?x=1", 0, Input{Name: "a", Value: "1"})
	if got := mustQuery(t, j, "http://example.com/p?y=2", 0); got != "a=1" {
		t.Fatalf("query string must not affect matching: %q", got)
	}
}

func TestOrderingLongestPathFirst(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "root", Value: "r", Path: "/"})
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "deep", Value: "d", Path: "/a/b"})
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "mid", Value: "m", Path: "/a"})
	got := mustQuery(t, j, "http://example.com/a/b/c", 0)
	if got != "deep=d; mid=m; root=r" {
		t.Fatalf("got %q", got)
	}
}

func TestSamePathCreationOrder(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "b", Value: "2"})
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "a", Value: "1"})
	if got := mustQuery(t, j, "http://example.com/", 0); got != "b=2; a=1" {
		t.Fatalf("creation order: %q", got)
	}
}

func TestReplaceKeepsCreationOrder(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "a", Value: "1"})
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "b", Value: "2"})
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "a", Value: "9"})
	if got := mustQuery(t, j, "http://example.com/", 0); got != "a=9; b=2" {
		t.Fatalf("replace must keep original position: %q", got)
	}
}

func TestSameNameDifferentPathsCoexist(t *testing.T) {
	j := New()
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "n", Value: "1", Path: "/"})
	mustReceive(t, j, "http://example.com/", 0, Input{Name: "n", Value: "2", Path: "/a"})
	if got := mustQuery(t, j, "http://example.com/a/x", 0); got != "n=2; n=1" {
		t.Fatalf("same-name cookies must not merge: %q", got)
	}
}
