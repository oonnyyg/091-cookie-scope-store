package jar

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Receive validates and stores (or deletes) a cookie at the given time.
func (j *Jar) Receive(rawURL string, now int64, in Input) error {
	_, host, reqPath, err := parseURL(rawURL)
	if err != nil {
		return err
	}
	if in.Name == "" {
		return errors.New("cookie name must not be empty")
	}
	if !isASCIIAlnum(in.Name) {
		return errors.New("cookie name must be ASCII letters or digits")
	}
	if !isASCIIAlnum(in.Value) {
		return errors.New("cookie value must be ASCII letters or digits")
	}

	c := &Cookie{
		Name:   in.Name,
		Value:  in.Value,
		Secure: in.Secure,
	}

	if in.Domain == "" {
		c.Domain = host
		c.HostOnly = true
	} else {
		d := strings.ToLower(strings.TrimPrefix(in.Domain, "."))
		if !isASCIIDomain(d) || net.ParseIP(d) != nil {
			return errors.New("invalid cookie domain")
		}
		if !domainMatch(host, d, false) {
			return fmt.Errorf("domain %q does not match origin host %q", in.Domain, host)
		}
		c.Domain = d
	}

	if in.Path == "" {
		c.Path = defaultPath(reqPath)
	} else {
		if !strings.HasPrefix(in.Path, "/") {
			return errors.New("cookie path must start with /")
		}
		c.Path = in.Path
	}

	if in.MaxAge != nil {
		if *in.MaxAge <= 0 {
			j.delete(c.Name, c.Domain, c.Path)
			return nil
		}
		c.Persistent = true
		c.ExpiresAt = now + *in.MaxAge
	}

	// Replace an identical (name, domain, path) record, keeping its
	// original creation order.
	for i, old := range j.cookies {
		if old.Name == c.Name && old.Domain == c.Domain && old.Path == c.Path {
			c.Created = old.Created
			j.cookies[i] = c
			return nil
		}
	}
	c.Created = j.nextID
	j.nextID++
	j.cookies = append(j.cookies, c)
	return nil
}

func (j *Jar) delete(name, domain, path string) {
	out := j.cookies[:0]
	for _, c := range j.cookies {
		if !(c.Name == name && c.Domain == domain && c.Path == path) {
			out = append(out, c)
		}
	}
	j.cookies = out
}
