package leaderboard

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type ScoreEntry struct {
	Name  string    `json:"name,omitempty"`
	Score int       `json:"score"`
	When  time.Time `json:"when"`
}

type Leaderboard struct {
	Scores []ScoreEntry `json:"scores"`
	path   string
	cap    int
}

func defaultLBPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "2nl-go", "scores.json")
}

// LoadLeaderboard always returns a usable lb. On read/parse errors it
// returns an empty one alongside the error — the first Save overwrites the
// damaged file (self-healing); the caller decides whether to log.
func LoadLeaderboard() (*Leaderboard, error) {
	lb := &Leaderboard{cap: 10, path: defaultLBPath(), Scores: []ScoreEntry{}}
	b, err := os.ReadFile(lb.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return lb, nil // first run, no scores yet
		}
		return lb, fmt.Errorf("read %s: %w", lb.path, err)
	}
	if err := json.Unmarshal(b, &lb.Scores); err != nil {
		lb.Scores = []ScoreEntry{} // discard partial decode
		return lb, fmt.Errorf("parse %s: %w", lb.path, err)
	}
	// trust no ordering in the file: Add/Qualifies rely on descending order
	slices.SortFunc(lb.Scores, func(a, b ScoreEntry) int { return cmp.Compare(b.Score, a.Score) })
	return lb, nil
}

func (lb *Leaderboard) Save() error {
	b, err := json.MarshalIndent(lb.Scores, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(lb.path), 0o755); err != nil {
		return err
	}
	tmp := lb.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, lb.path)
}

// Qualifies reports whether score would make the leaderboard.
func (lb *Leaderboard) Qualifies(score int) bool {
	if score <= 0 {
		return false
	}
	return len(lb.Scores) < lb.cap || score > lb.Scores[len(lb.Scores)-1].Score
}

// Add inserts score and returns its 0-based rank, or -1 if it was rejected
// or did not make the cut
func (lb *Leaderboard) Add(score int, name string) int {
	if !lb.Qualifies(score) {
		return -1
	}
	idx := len(lb.Scores)
	for i, e := range lb.Scores {
		if e.Score < score {
			idx = i
			break
		}
	}
	lb.Scores = slices.Insert(lb.Scores, idx, ScoreEntry{Name: name, Score: score, When: time.Now()})
	if len(lb.Scores) > lb.cap {
		lb.Scores = lb.Scores[:lb.cap]
	}
	return idx
}
