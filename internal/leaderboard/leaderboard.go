package leaderboard

import (
	"cmp"
	"encoding/json"
	"fmt"
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
	dir, _ := os.UserConfigDir()
	if dir == "" {
		dir = "."
	}
	dir = filepath.Join(dir, "2nl-go")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "scores.json")
}

func LoadLeaderboard() *Leaderboard {
	lb := &Leaderboard{cap: 10, path: defaultLBPath(), Scores: []ScoreEntry{}}
	if b, err := os.ReadFile(lb.path); err == nil {
		_ = json.Unmarshal(b, &lb.Scores)
	}
	return lb
}

func (lb *Leaderboard) Save() {
	b, _ := json.MarshalIndent(lb.Scores, "", "  ")
	_ = os.MkdirAll(filepath.Dir(lb.path), 0o755)
	tmp := lb.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		fmt.Println("lb write tmp:", err)
		_ = os.WriteFile(lb.path, b, 0o600)
		return
	}
	if err := os.Rename(tmp, lb.path); err != nil {
		fmt.Println("lb rename:", err)
		_ = os.WriteFile(lb.path, b, 0o600)
		return
	}
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
	if score <= 0 {
		return -1
	}
	when := time.Now()
	lb.Scores = append(lb.Scores, ScoreEntry{Name: name, Score: score, When: when})
	slices.SortFunc(lb.Scores, func(a, b ScoreEntry) int { return cmp.Compare(b.Score, a.Score) })
	if len(lb.Scores) > lb.cap {
		lb.Scores = lb.Scores[:lb.cap]
	}
	return slices.IndexFunc(lb.Scores, func(e ScoreEntry) bool { return e.When.Equal(when) })
}
