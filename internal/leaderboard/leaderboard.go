package leaderboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ScoreEntry struct {
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
	} else if !os.IsNotExist(err) {
	}
	return lb
}

func (lb *Leaderboard) Save() {
	b, _ := json.MarshalIndent(lb.Scores, "", "  ")
	_ = os.MkdirAll(filepath.Dir(lb.path), 0o755)
	tmp := lb.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		fmt.Println("lb write tmp:", err)
		_ = os.WriteFile(lb.path, b, 0o644)
		return
	}
	if err := os.Rename(tmp, lb.path); err != nil {
		fmt.Println("lb rename:", err)
		_ = os.WriteFile(lb.path, b, 0o644)
		return
	}
}

func (lb *Leaderboard) Add(score int) {
	if score <= 0 {
		return
	}
	lb.Scores = append(lb.Scores, ScoreEntry{Score: score, When: time.Now()})
	sort.Slice(lb.Scores, func(i, j int) bool { return lb.Scores[i].Score > lb.Scores[j].Score })
	if len(lb.Scores) > lb.cap {
		lb.Scores = lb.Scores[:lb.cap]
	}
}
