package logic

import (
	"math/rand"
	"time"
)

type Config struct {
	NumColumns, NumRows           int
	StartingColumn, StartingRow   int
	PreviewColumn, PreviewRow     int
	SpawnStart, SpawnEnd          int
	PointsPerLine, LevelThreshold int
	RNG                           *rand.Rand
}

func DefaultConfig() Config {
	return Config{
		NumColumns:     10,
		NumRows:        10,
		StartingColumn: 4,
		StartingRow:    4,
		PreviewColumn:  4,
		PreviewRow:     12,
		SpawnStart:     3,
		SpawnEnd:       6,
		PointsPerLine:  50,
		LevelThreshold: 1000,
		RNG:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
