package logic

type TileType struct {
	ID     int
	Column int
	Row    int
	Tile   int
}

type Line struct {
	Tiles []*TileType
	Type  LineType
}

func NewLine(t LineType) *Line {
	return &Line{Type: t}
}

func (l *Line) Add(t *TileType) { l.Tiles = append(l.Tiles, t) }
func (l *Line) Length() int     { return len(l.Tiles) }

type LineType int

const (
	Horizontal LineType = iota
	Vertical
	Diagonal
)
