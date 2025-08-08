package logic

import "math/rand"

type Orientation int

const (
	Zero Orientation = iota
	Ninety
	OneEighty
	TwoSeventy
)

func rotate(o Orientation, cw bool) Orientation {
	if cw {
		if o == TwoSeventy {
			return Zero
		}
		return o + 1
	}
	if o == Zero {
		return TwoSeventy
	}
	return o - 1
}

type pt struct{ dx, dy int }

type Shape struct {
	Column, Row int
	Orientation Orientation
	Blocks      []TileType
	pos         map[Orientation][]pt
}

func newShape(column, row int, pos map[Orientation][]pt, rng *rand.Rand) *Shape {
	s := &Shape{Column: column, Row: row, Orientation: Orientation(rng.Intn(4)), pos: pos}
	s.initializeBlocks(rng)
	return s
}

func (s *Shape) initializeBlocks(rng *rand.Rand) {
	arr := s.pos[s.Orientation]
	s.Blocks = make([]TileType, len(arr))
	for i, d := range arr {
		s.Blocks[i] = newTile(s.Column+d.dx, s.Row+d.dy, randomForTile(rng))
	}
}

func (s *Shape) rotateBlocks(o Orientation) {
	arr := s.pos[o]
	for i, d := range arr {
		s.Blocks[i].Column = s.Column + d.dx
		s.Blocks[i].Row = s.Row + d.dy
	}
}

func (s *Shape) RotateClockwise() {
	no := rotate(s.Orientation, true)
	s.rotateBlocks(no)
	s.Orientation = no
}

func (s *Shape) RotateCounterClockwise() {
	no := rotate(s.Orientation, false)
	s.rotateBlocks(no)
	s.Orientation = no
}

func (s *Shape) ShiftBy(dc, dr int) {
	s.Column += dc
	s.Row += dr
	for i := range s.Blocks {
		s.Blocks[i].Column += dc
		s.Blocks[i].Row += dr
	}
}

func (s *Shape) MoveTo(c, r int) {
	s.Column, s.Row = c, r
	s.rotateBlocks(s.Orientation)
}

func randomForTile(rng *rand.Rand) int { return rng.Intn(6) + 1 }

var (
	oneBlock = map[Orientation][]pt{
		Zero: {{0, 0}}, Ninety: {{0, 0}}, OneEighty: {{0, 0}}, TwoSeventy: {{0, 0}},
	}
	twoBlock = map[Orientation][]pt{
		Zero: {{0, 1}, {1, 1}}, Ninety: {{1, 1}, {1, 0}},
		OneEighty: {{1, 0}, {0, 0}}, TwoSeventy: {{0, 0}, {0, 1}},
	}
	threeLine = map[Orientation][]pt{
		Zero: {{-1, 0}, {0, 0}, {1, 0}}, Ninety: {{0, 1}, {0, 0}, {0, -1}},
		OneEighty: {{1, 0}, {0, 0}, {-1, 0}}, TwoSeventy: {{0, -1}, {0, 0}, {0, 1}},
	}
	line4 = map[Orientation][]pt{
		Zero:       {{-1, 0}, {0, 0}, {1, 0}, {2, 0}},
		Ninety:     {{0, -1}, {0, 0}, {0, 1}, {0, 2}},
		OneEighty:  {{2, 0}, {1, 0}, {0, 0}, {-1, 0}},
		TwoSeventy: {{0, 2}, {0, 1}, {0, 0}, {0, -1}},
	}
	square = map[Orientation][]pt{
		Zero:       {{0, 0}, {0, 1}, {1, 1}, {1, 0}},
		OneEighty:  {{1, 1}, {1, 0}, {0, 0}, {0, 1}},
		Ninety:     {{0, 1}, {1, 1}, {1, 0}, {0, 0}},
		TwoSeventy: {{1, 0}, {0, 0}, {0, 1}, {1, 1}},
	}
	tshape = map[Orientation][]pt{
		Zero:       {{0, 1}, {0, 0}, {0, -1}, {1, 0}},
		Ninety:     {{1, 0}, {0, 0}, {-1, 0}, {0, -1}},
		OneEighty:  {{0, -1}, {0, 0}, {0, 1}, {-1, 0}},
		TwoSeventy: {{-1, 0}, {0, 0}, {1, 0}, {0, 1}},
	}
	lshape = map[Orientation][]pt{
		Zero:       {{0, 1}, {0, 0}, {0, -1}, {1, -1}},
		Ninety:     {{1, 0}, {0, 0}, {-1, 0}, {-1, -1}},
		OneEighty:  {{0, -1}, {0, 0}, {0, 1}, {-1, 1}},
		TwoSeventy: {{-1, 0}, {0, 0}, {1, 0}, {1, 1}},
	}
	jshape = map[Orientation][]pt{
		Zero:       {{0, 1}, {0, 0}, {0, -1}, {-1, -1}},
		Ninety:     {{1, 0}, {0, 0}, {-1, 0}, {-1, 1}},
		OneEighty:  {{0, -1}, {0, 0}, {0, 1}, {1, 1}},
		TwoSeventy: {{-1, 0}, {0, 0}, {1, 0}, {1, -1}},
	}
	zshape = map[Orientation][]pt{
		Zero:       {{0, 1}, {0, 0}, {-1, 0}, {-1, -1}},
		Ninety:     {{1, -1}, {0, -1}, {0, 0}, {-1, 0}},
		OneEighty:  {{-1, -1}, {-1, 0}, {0, 0}, {0, 1}},
		TwoSeventy: {{-1, 0}, {0, 0}, {0, -1}, {1, -1}},
	}
	sshape = map[Orientation][]pt{
		Zero:       {{-1, -1}, {-1, 0}, {0, 0}, {0, 1}},
		Ninety:     {{1, 0}, {0, 0}, {0, 1}, {-1, 1}},
		OneEighty:  {{0, 1}, {0, 0}, {-1, 0}, {-1, -1}},
		TwoSeventy: {{-1, 1}, {0, 1}, {0, 0}, {1, 0}},
	}
	jshort = map[Orientation][]pt{
		Zero:       {{0, 1}, {0, 0}, {1, 0}},
		Ninety:     {{1, 0}, {0, 0}, {0, -1}},
		OneEighty:  {{0, -1}, {0, 0}, {-1, 0}},
		TwoSeventy: {{-1, 0}, {0, 0}, {0, 1}},
	}
)
