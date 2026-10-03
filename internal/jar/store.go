package jar

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Store wraps a Jar with JSON-file persistence and monotonic time.
type Store struct {
	jar      *Jar
	dir      string
	lastTime int64
}

type diskState struct {
	LastTime int64     `json:"last_time"`
	NextID   int64     `json:"next_id"`
	Cookies  []*Cookie `json:"cookies"`
}

func (s *Store) statePath() string { return filepath.Join(s.dir, "state.json") }

// OpenStore loads persistent cookies from dir, creating it if needed.
func OpenStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{jar: New(), dir: dir}
	data, err := os.ReadFile(s.statePath())
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var st diskState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("corrupt state file: %w", err)
	}
	s.lastTime = st.LastTime
	s.jar.nextID = st.NextID
	for _, c := range st.Cookies {
		if c.Persistent {
			s.jar.cookies = append(s.jar.cookies, c)
		}
	}
	return s, nil
}

// CheckTime enforces that time never goes backwards, across restarts too.
func (s *Store) CheckTime(t int64) error {
	if t < 0 {
		return errors.New("time must be a non-negative integer")
	}
	if t < s.lastTime {
		return fmt.Errorf("time %d goes backwards (last %d)", t, s.lastTime)
	}
	s.lastTime = t
	return nil
}

// Receive applies a receive operation and persists the result.
func (s *Store) Receive(rawURL string, now int64, in Input) error {
	if err := s.jar.Receive(rawURL, now, in); err != nil {
		return err
	}
	return s.save()
}

// Query applies a query operation and persists any expiry cleanup.
func (s *Store) Query(rawURL string, now int64) (string, error) {
	h, err := s.jar.Query(rawURL, now)
	if err != nil {
		return "", err
	}
	return h, s.save()
}

func (s *Store) save() error {
	st := diskState{LastTime: s.lastTime, NextID: s.jar.nextID}
	for _, c := range s.jar.cookies {
		if c.Persistent {
			st.Cookies = append(st.Cookies, c)
		}
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.statePath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.statePath())
}
