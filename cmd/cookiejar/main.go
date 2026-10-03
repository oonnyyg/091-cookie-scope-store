// Command cookiejar is a line-delimited JSON cookie store.
//
// Usage: go run ./cmd/cookiejar --state DIR
//
// Input (one JSON object per line):
//
//	{"op":"receive","url":"https://a.example.com/x/y","time":100,
//	 "cookie":{"name":"sid","value":"abc","domain":".example.com",
//	           "path":"/x","secure":true,"max_age":3600}}
//	{"op":"query","url":"https://a.example.com/x/y/z","time":120}
//
// Output (one JSON object per line):
//
//	{"ok":true}                        receive succeeded
//	{"ok":true,"header":"sid=abc"}     query result (Cookie header)
//	{"ok":false,"error":"..."}         any failure; store unchanged
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"example.com/cookie-scope-store/internal/jar"
)

type request struct {
	Op     string     `json:"op"`
	URL    string     `json:"url"`
	Time   *int64     `json:"time"`
	Cookie *jar.Input `json:"cookie"`
}

type response struct {
	OK     bool   `json:"ok"`
	Header string `json:"header,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	stateDir := flag.String("state", "", "state directory (required)")
	flag.Parse()
	if *stateDir == "" {
		fmt.Fprintln(os.Stderr, "missing --state directory")
		os.Exit(2)
	}
	store, err := jar.OpenStore(*stateDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open state:", err)
		os.Exit(1)
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	enc := json.NewEncoder(out)

	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		enc.Encode(handle(store, line))
		out.Flush()
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "read stdin:", err)
		os.Exit(1)
	}
}

func handle(store *jar.Store, line []byte) response {
	var req request
	if err := json.Unmarshal(line, &req); err != nil {
		return response{Error: "invalid JSON: " + err.Error()}
	}
	if req.Time == nil {
		return response{Error: "missing time"}
	}
	if err := store.CheckTime(*req.Time); err != nil {
		return response{Error: err.Error()}
	}
	switch req.Op {
	case "receive":
		if req.Cookie == nil {
			return response{Error: "missing cookie"}
		}
		if err := store.Receive(req.URL, *req.Time, *req.Cookie); err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true}
	case "query":
		header, err := store.Query(req.URL, *req.Time)
		if err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true, Header: header}
	default:
		return response{Error: "unknown op " + fmt.Sprintf("%q", req.Op)}
	}
}
