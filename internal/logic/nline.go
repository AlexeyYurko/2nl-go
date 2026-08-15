package logic

type NLineDelegate interface {
	GameDidEnd(*NLine)
	GameDidBegin(*NLine)
	GameShapeDidLand(*NLine)
	GameShapeDidMove(*NLine)
	GameShapeDidDrop(*NLine)
	ColorShiftMake(*NLine)
	GameDidLevelUp(*NLine)
	LinesMatched(*NLine, []*Line)
}

type NLine struct {
	cfg          Config
	BlockArray   [][]*TileType
	NextShape    *Shape
	FallingShape *Shape
	Delegate     NLineDelegate

	Score int
	Round int

	numShapeTypes int
}

func NewNLineDefault() *NLine { return NewNLine(DefaultConfig()) }

func NewNLine(cfg Config) *NLine {
	a := make([][]*TileType, cfg.NumColumns)
	for i := range a {
		a[i] = make([]*TileType, cfg.NumRows)
	}
	return &NLine{
		cfg:           cfg,
		BlockArray:    a,
		Round:         1,
		numShapeTypes: 4,
	}
}

func (n *NLine) BeginGame() {
	n.clearAll()
	n.Score = 0
	n.Round = 1
	n.roundShapeCap()
	if n.NextShape == nil {
		n.NextShape = n.randomShape(n.cfg.PreviewColumn, n.cfg.PreviewRow)
	}
	if n.Delegate != nil {
		n.Delegate.GameDidBegin(n)
	}
}

func (n *NLine) NewShape() (falling, next *Shape) {
	n.FallingShape = n.NextShape
	n.NextShape = n.randomShape(n.cfg.PreviewColumn, n.cfg.PreviewRow)
	n.FallingShape.MoveTo(n.cfg.StartingColumn, n.cfg.StartingRow)

	if !n.detectPossibleSpawnPoint() {
		n.NextShape = n.FallingShape
		n.NextShape.MoveTo(n.cfg.PreviewColumn, n.cfg.PreviewRow)
		n.endGame()
		return nil, nil
	}
	return n.FallingShape, n.NextShape
}

var shapeDefs = []map[Orientation][]pt{
	oneBlock, twoBlock, threeLine, jshort, square,
	line4, tshape, lshape, jshape, sshape, zshape,
}

func (n *NLine) randomShape(c, r int) *Shape {
	return newShape(c, r, shapeDefs[n.cfg.RNG.Intn(n.numShapeTypes)], n.cfg.RNG)
}

// Spawn offset logic
func (n *NLine) detectPossibleSpawnPoint() bool {
	type xy struct{ x, y int }
	diffs := []xy{}
	for dy := -1; dy <= 2; dy++ {
		for dx := -1; dx <= 2; dx++ {
			if n.detectPlacement(dx, dy) {
				diffs = append(diffs, xy{dx, dy})
			}
		}
	}
	if len(diffs) > 0 {
		off := diffs[n.cfg.RNG.Intn(len(diffs))]
		s := n.FallingShape
		for i := range s.Blocks {
			s.Blocks[i].Column += off.x
			s.Blocks[i].Row += off.y
		}
		s.Column += off.x
		s.Row += off.y
		n.FallingShape = s
		return true
	}
	return false
}

func (n *NLine) detectPlacement(dx, dy int) bool {
	s := n.FallingShape
	if s == nil {
		return false
	}
	for _, b := range s.Blocks {
		c := b.Column + dx
		r := b.Row + dy
		if c < n.cfg.SpawnStart || c > n.cfg.SpawnEnd || r < n.cfg.SpawnStart || r > n.cfg.SpawnEnd {
			return false
		}
		if n.BlockArray[c][r] != nil {
			return false
		}
	}
	return true
}

func (n *NLine) detectIllegalPlacement() bool {
	s := n.FallingShape
	if s == nil {
		return false
	}
	for _, b := range s.Blocks {
		if b.Column < 0 || b.Column >= n.cfg.NumColumns || b.Row < 0 || b.Row >= n.cfg.NumRows {
			return true
		}
		if n.BlockArray[b.Column][b.Row] != nil {
			return true
		}
	}
	return false
}

func (n *NLine) settleShape() {
	s := n.FallingShape
	if s == nil {
		return
	}
	for i := range s.Blocks {
		b := s.Blocks[i]
		cp := b
		n.BlockArray[b.Column][b.Row] = &cp
	}
	n.FallingShape = nil
	if n.Delegate != nil {
		n.Delegate.GameShapeDidLand(n)
	}
}

func (n *NLine) endGame() {
	if n.Delegate != nil {
		n.Delegate.GameDidEnd(n)
	}
	n.Score = 0
	n.Round = 1
}

func (n *NLine) DropShape() {
	if !n.detectIllegalPlacement() {
		if n.Delegate != nil {
			n.Delegate.GameShapeDidDrop(n)
		}
		return
	}
	n.endGame()
}

func (n *NLine) LetShapeFall() {
	if n.detectIllegalPlacement() {
		n.endGame()
	} else {
		n.settleShape()
		lines := n.removeMatches()
		if len(lines) > 0 {
			for range lines {
				n.scoreForRemove()
			}
			if n.Delegate != nil {
				n.Delegate.LinesMatched(n, lines)
			}
		}
	}
}

func (n *NLine) RotateShape() {
	s := n.FallingShape
	if s == nil || len(s.Blocks) == 1 {
		return
	}
	s.RotateClockwise()
	if n.detectIllegalPlacement() {
		s.RotateCounterClockwise()
		return
	}
	if n.Delegate != nil {
		n.Delegate.GameShapeDidMove(n)
	}
}

func (n *NLine) MoveShapeLeft() {
	s := n.FallingShape
	if s == nil {
		return
	}
	s.ShiftBy(-1, 0)
	if n.detectIllegalPlacement() {
		s.ShiftBy(1, 0)
		return
	}
	if n.Delegate != nil {
		n.Delegate.GameShapeDidMove(n)
	}
}

func (n *NLine) MoveShapeRight() {
	s := n.FallingShape
	if s == nil {
		return
	}
	s.ShiftBy(1, 0)
	if n.detectIllegalPlacement() {
		s.ShiftBy(-1, 0)
		return
	}
	if n.Delegate != nil {
		n.Delegate.GameShapeDidMove(n)
	}
}

func (n *NLine) MoveShapeUp() {
	s := n.FallingShape
	if s == nil {
		return
	}
	s.ShiftBy(0, -1)
	if n.detectIllegalPlacement() {
		s.ShiftBy(0, 1)
		return
	}
	if n.Delegate != nil {
		n.Delegate.GameShapeDidMove(n)
	}
}

func (n *NLine) MoveShapeDown() {
	s := n.FallingShape
	if s == nil {
		return
	}
	s.ShiftBy(0, 1)
	if n.detectIllegalPlacement() {
		s.ShiftBy(0, -1)
		return
	}
	if n.Delegate != nil {
		n.Delegate.GameShapeDidMove(n)
	}
}

func (n *NLine) ColorShift() {
	s := n.FallingShape
	if s == nil {
		return
	}
	if len(s.Blocks) <= 1 {
		return
	}
	tmp := make([]int, len(s.Blocks))
	for i := range s.Blocks {
		tmp[i] = s.Blocks[i].Tile
	}
	for i := 0; i < len(s.Blocks)-1; i++ {
		s.Blocks[i].Tile = tmp[i+1]
	}
	s.Blocks[len(s.Blocks)-1].Tile = tmp[0]
	if n.Delegate != nil {
		n.Delegate.ColorShiftMake(n)
	}
}

func (n *NLine) clearAll() {
	for r := 0; r < n.cfg.NumRows; r++ {
		for c := 0; c < n.cfg.NumColumns; c++ {
			n.BlockArray[c][r] = nil
		}
	}
	n.FallingShape = nil
	n.NextShape = nil
}

// scoring / level
func (n *NLine) scoreForRemove() {
	points := n.cfg.PointsPerLine * n.Round
	n.Score += points
	if n.Score >= n.Round*n.cfg.LevelThreshold {
		n.Round++
		n.roundShapeCap()
		if n.Delegate != nil {
			n.Delegate.GameDidLevelUp(n)
		}
	}
}

func (n *NLine) roundShapeCap() {
	switch {
	case n.Round <= 2:
		n.numShapeTypes = 4
	case n.Round <= 4:
		n.numShapeTypes = 7
	default:
		n.numShapeTypes = 11
	}
}

// matching
func (n *NLine) removeMatches() []*Line {
	hs := n.detectHorizontalMatches()
	vs := n.detectVerticalMatches()
	ds := n.detectDiagonalMatches()
	all := make([]*Line, 0, len(hs)+len(vs)+len(ds))
	all = append(all, hs...)
	all = append(all, vs...)
	all = append(all, ds...)
	// clear blocks
	for _, ln := range all {
		for _, t := range ln.Tiles {
			n.BlockArray[t.Column][t.Row] = nil
		}
	}
	return all
}

func (n *NLine) detectHorizontalMatches() []*Line {
	var res []*Line
	for r := 0; r < n.cfg.NumRows; r++ {
		c := 0
		for c <= n.cfg.NumColumns-3 {
			t := n.BlockArray[c][r]
			if t != nil {
				mt := t.Tile
				if n.BlockArray[c+1][r] != nil && n.BlockArray[c+1][r].Tile == mt &&
					n.BlockArray[c+2][r] != nil && n.BlockArray[c+2][r].Tile == mt {
					ln := NewLine(Horizontal)
					for c < n.cfg.NumColumns && n.BlockArray[c][r] != nil && n.BlockArray[c][r].Tile == mt {
						ln.Add(n.BlockArray[c][r])
						c++
					}
					res = append(res, ln)
					continue
				}
			}
			c++
		}
	}
	return res
}

func (n *NLine) detectVerticalMatches() []*Line {
	var res []*Line
	for c := 0; c < n.cfg.NumColumns; c++ {
		r := 0
		for r <= n.cfg.NumRows-3 {
			t := n.BlockArray[c][r]
			if t != nil {
				mt := t.Tile
				if n.BlockArray[c][r+1] != nil && n.BlockArray[c][r+1].Tile == mt &&
					n.BlockArray[c][r+2] != nil && n.BlockArray[c][r+2].Tile == mt {
					ln := NewLine(Vertical)
					for r < n.cfg.NumRows && n.BlockArray[c][r] != nil && n.BlockArray[c][r].Tile == mt {
						ln.Add(n.BlockArray[c][r])
						r++
					}
					res = append(res, ln)
					continue
				}
			}
			r++
		}
	}
	return res
}

func (n *NLine) detectDiagonalMatches() []*Line {
	var res []*Line
	// /
	for row := 0; row <= n.cfg.NumRows-3; row++ {
		for col := 0; col <= n.cfg.NumColumns-3; col++ {
			if ln := n.collectRun(col, row, 1, 1); ln != nil {
				res = append(res, ln)
			}
		}
	}
	// \
	for row := 0; row <= n.cfg.NumRows-3; row++ {
		for col := n.cfg.NumColumns - 1; col >= 2; col-- {
			if ln := n.collectRun(col, row, -1, 1); ln != nil {
				res = append(res, ln)
			}
		}
	}
	return res
}

func (n *NLine) collectRun(c, r, dc, dr int) *Line {
	if !n.inBounds(c, r) || !n.inBounds(c+dc, r+dr) || !n.inBounds(c+2*dc, r+2*dr) {
		return nil
	}
	t := n.BlockArray[c][r]
	if t == nil {
		return nil
	}
	mt := t.Tile
	if !n.sameTile(c+dc, r+dr, mt) || !n.sameTile(c+2*dc, r+2*dr, mt) {
		return nil
	}
	ln := NewLine(Diagonal)
	for n.inBounds(c, r) && n.sameTile(c, r, mt) {
		ln.Add(n.BlockArray[c][r])
		c += dc
		r += dr
	}
	return ln
}

func (n *NLine) sameTile(c, r, tile int) bool {
	t := n.BlockArray[c][r]
	return t != nil && t.Tile == tile
}

func (n *NLine) inBounds(c, r int) bool {
	return c >= 0 && c < n.cfg.NumColumns && r >= 0 && r < n.cfg.NumRows
}
