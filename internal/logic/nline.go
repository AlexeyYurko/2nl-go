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
	nextID        int
	over          bool
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
	n.over = false
	n.roundShapeCap()
	if n.NextShape == nil {
		n.NextShape = n.randomShape(n.cfg.PreviewColumn, n.cfg.PreviewRow)
	}
	if n.Delegate != nil {
		n.Delegate.GameDidBegin(n)
	}
}

// Over reports whether the game has ended. All mutating methods are no-ops
// after that, until BeginGame starts a new game.
func (n *NLine) Over() bool { return n.over }

func (n *NLine) newTile(column, row, tile int) TileType {
	n.nextID++
	return TileType{ID: n.nextID, Column: column, Row: row, Tile: tile}
}

func (n *NLine) NewShape() (falling, next *Shape) {
	if n.over {
		return nil, nil
	}
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
	return newShape(n, c, r, shapeDefs[n.cfg.RNG.Intn(n.numShapeTypes)], n.cfg.RNG)
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
	if n.over {
		return
	}
	n.over = true
	if n.Delegate != nil {
		n.Delegate.GameDidEnd(n)
	}
	n.Score = 0
	n.Round = 1
}

func (n *NLine) DropShape() {
	if n.over {
		return
	}
	if !n.detectIllegalPlacement() {
		if n.Delegate != nil {
			n.Delegate.GameShapeDidDrop(n)
		}
		return
	}
	n.endGame()
}

func (n *NLine) LetShapeFall() {
	if n.over {
		return
	}
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
	if n.over {
		return
	}
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

func (n *NLine) moveShape(dc, dr int) {
	if n.over {
		return
	}
	s := n.FallingShape
	if s == nil {
		return
	}
	s.ShiftBy(dc, dr)
	if n.detectIllegalPlacement() {
		s.ShiftBy(-dc, -dr)
		return
	}
	if n.Delegate != nil {
		n.Delegate.GameShapeDidMove(n)
	}
}

func (n *NLine) MoveShapeLeft()  { n.moveShape(-1, 0) }
func (n *NLine) MoveShapeRight() { n.moveShape(1, 0) }
func (n *NLine) MoveShapeUp()    { n.moveShape(0, -1) }
func (n *NLine) MoveShapeDown()  { n.moveShape(0, 1) }

func (n *NLine) ColorShift() {
	if n.over {
		return
	}
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
var runDirs = [4]struct {
	dc, dr int
	kind   LineType
}{
	{1, 0, Horizontal},
	{0, 1, Vertical},
	{1, 1, Diagonal},
	{-1, 1, Diagonal},
}

func (n *NLine) removeMatches() []*Line {
	all := make([]*Line, 0, 4)
	for _, d := range runDirs {
		all = append(all, n.detectRuns(d.dc, d.dr, d.kind)...)
	}
	// clear blocks
	for _, ln := range all {
		for _, t := range ln.Tiles {
			n.BlockArray[t.Column][t.Row] = nil
		}
	}
	return all
}

// detectRuns collects maximal runs of 3+ same-tile blocks in direction
// (dc, dr). Each run is reported exactly once, starting from its first tile
// (the one with no same-tile predecessor), so overlapping sub-lines of a
// long run are not double-counted.
func (n *NLine) detectRuns(dc, dr int, kind LineType) []*Line {
	var res []*Line
	for r := 0; r < n.cfg.NumRows; r++ {
		for c := 0; c < n.cfg.NumColumns; c++ {
			t := n.BlockArray[c][r]
			if t == nil {
				continue
			}
			if n.sameTile(c-dc, r-dr, t.Tile) {
				continue // mid-run, already reported
			}
			if !n.sameTile(c+dc, r+dr, t.Tile) || !n.sameTile(c+2*dc, r+2*dr, t.Tile) {
				continue
			}
			ln := NewLine(kind)
			for cc, rr := c, r; n.sameTile(cc, rr, t.Tile); cc, rr = cc+dc, rr+dr {
				ln.Add(n.BlockArray[cc][rr])
			}
			res = append(res, ln)
		}
	}
	return res
}

func (n *NLine) sameTile(c, r, tile int) bool {
	return n.inBounds(c, r) && n.BlockArray[c][r] != nil && n.BlockArray[c][r].Tile == tile
}

func (n *NLine) inBounds(c, r int) bool {
	return c >= 0 && c < n.cfg.NumColumns && r >= 0 && r < n.cfg.NumRows
}
