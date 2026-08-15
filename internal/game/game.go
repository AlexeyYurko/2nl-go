package game

import (
	"2nline/internal/assets"
	"2nline/internal/leaderboard"
	"2nline/internal/logic"
	"2nline/internal/sfx"
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

type Anim struct {
	x0, y0, x1, y1 float64
	t, d           float64
	s0, s1         float64
	dead           bool
}

func (a *Anim) step(dt float64) {
	a.t += dt
	if a.t >= a.d {
		a.t = a.d
		a.dead = true
	}
}
func (a *Anim) pos() (x, y, s float64) {
	if a.d <= 0 {
		return a.x1, a.y1, a.s1
	}
	u := a.t / a.d
	if u < 0 {
		u = 0
	}
	if u > 1 {
		u = 1
	}
	// easeOutQuad
	u = 1 - (1-u)*(1-u)
	x = a.x0 + (a.x1-a.x0)*u
	y = a.y0 + (a.y1-a.y0)*u
	s = a.s0 + (a.s1-a.s0)*u
	return x, y, s
}

type Ghost struct {
	name string
	anim *Anim
}
type Game struct {
	nl *logic.NLine

	atlas *assets.Atlas

	cfg        logic.Config
	tileSize   int
	scale      float64 // for @2x assets
	windowW    int
	windowH    int
	layerOffX  float64
	layerOffY  float64
	isPaused   bool
	timeLeft   float64 // seconds to tick
	tickLength float64 // seconds

	anims          map[int]*Anim
	lastVis        map[int][2]float64
	ghosts         []*Ghost
	settleTimer    float64
	nextShapeTimer float64

	timerRadius float32
	timerPosX   float32
	timerPosY   float32

	sfx *sfx.SFX

	lb *leaderboard.Leaderboard

	tileOp *ebiten.DrawImageOptions
	textOp *text.DrawOptions

	hudScoreVal int
	hudScoreStr string
	hudLevelVal int
	hudLevelStr string
	hudLB       []string
	lbDirty     bool

	lastTS    time.Time
	logicAcc  float64
	logicStep float64

	panelScore *ebiten.Image
	panelLB    *ebiten.Image
}

func New() (*Game, error) {
	atlas, err := assets.LoadAtlas()
	if err != nil {
		return nil, err
	}
	g := &Game{
		atlas:      atlas,
		tileSize:   30,
		scale:      0.5, // assets are @2x ~60px
		windowW:    480,
		windowH:    640,
		tickLength: 15.0,
	}

	g.anims = map[int]*Anim{}
	g.lastVis = map[int][2]float64{}

	s, err := sfx.LoadSFX()
	if err != nil {
		return nil, err
	}
	g.sfx = s

	g.lb = leaderboard.LoadLeaderboard()

	g.tileOp = &ebiten.DrawImageOptions{}
	g.textOp = &text.DrawOptions{}
	g.lbDirty = true
	g.hudScoreVal = -1
	g.hudLevelVal = -1

	g.panelScore = ebiten.NewImage(170, 40)
	g.panelScore.Fill(color.NRGBA{0x00, 0x00, 0x00, 0x80})
	g.panelLB = ebiten.NewImage(170, 22+5*16)
	g.panelLB.Fill(color.NRGBA{0x00, 0x00, 0x00, 0x80})

	g.cfg = logic.DefaultConfig()

	g.layerOffX = float64(g.windowW)/2 - (float64(g.cfg.NumColumns)*float64(g.tileSize)+float64(g.cfg.NumColumns))/2
	g.layerOffY = float64(g.windowH)/2 - (float64(g.cfg.NumRows)*float64(g.tileSize)+float64(g.cfg.NumRows))/2

	g.timerRadius = 40
	g.timerPosX = float32(g.windowW) - 70
	g.timerPosY = 70

	g.tickLength = 15.0
	g.logicStep = 1.0 / 60.0
	g.lastTS = time.Now()

	nl := logic.NewNLineDefault()
	nl.Delegate = g
	g.nl = nl
	g.BeginGame()
	return g, nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (w, h int) {
	return g.windowW, g.windowH
}
func (g *Game) WindowW() int { return g.windowW }
func (g *Game) WindowH() int { return g.windowH }

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.isPaused = !g.isPaused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.BeginGame()
	}
	if g.isPaused {
		g.lastTS = time.Now() // to avoid huge dt after pause
		return nil
	}

	now := time.Now()
	dt := now.Sub(g.lastTS).Seconds()
	g.lastTS = now
	if dt > 0.05 { // clamp for lag
		dt = 0.05
	}

	g.handleInput()
	g.updateAnims(dt)
	g.updateLogic(dt)

	g.sfx.Update()
	return nil
}

func (g *Game) handleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.nl.RotateShape()
	}
	if keyRepeat(ebiten.KeyLeft) {
		g.nl.MoveShapeLeft()
	}
	if keyRepeat(ebiten.KeyRight) {
		g.nl.MoveShapeRight()
	}
	if keyRepeat(ebiten.KeyDown) {
		g.nl.MoveShapeDown()
	}
	if keyRepeat(ebiten.KeyUp) {
		g.nl.MoveShapeUp()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.nl.DropShape()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.nl.ColorShift()
	}
}

func (g *Game) updateAnims(dt float64) {
	for id, a := range g.anims {
		a.step(dt)
		if a.dead {
			delete(g.anims, id)
		}
	}
	for i := 0; i < len(g.ghosts); {
		g.ghosts[i].anim.step(dt)
		if g.ghosts[i].anim.dead {
			g.ghosts[i] = g.ghosts[len(g.ghosts)-1]
			g.ghosts = g.ghosts[:len(g.ghosts)-1]
		} else {
			i++
		}
	}
}

func (g *Game) updateLogic(dt float64) {
	// game logic/timers — fixed step
	g.logicAcc += dt
	for g.logicAcc >= g.logicStep {
		if g.settleTimer > 0 {
			g.settleTimer -= g.logicStep
			if g.settleTimer <= 0 {
				g.nl.LetShapeFall()
			}
		}
		if g.nextShapeTimer > 0 {
			g.nextShapeTimer -= g.logicStep
			if g.nextShapeTimer <= 0 {
				g.NextShape()
				g.nextShapeTimer = 0
			}
		}
		if g.settleTimer <= 0 && g.nextShapeTimer <= 0 {
			if g.timeLeft <= 0 {
				g.nl.DropShape()
			} else {
				g.timeLeft -= g.logicStep
			}
		}
		g.logicAcc -= g.logicStep
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawField(screen)
	g.drawSettled(screen)
	g.drawFallingShape(screen)
	g.drawGhosts(screen)
	g.drawNextShape(screen)
	g.drawHUD(screen)
}

func (g *Game) drawField(screen *ebiten.Image) {
	// Field (free/spawn)
	for r := 0; r < g.cfg.NumRows; r++ {
		for c := 0; c < g.cfg.NumColumns; c++ {
			spawn := (r >= g.cfg.SpawnStart && r <= g.cfg.SpawnEnd) && (c >= g.cfg.SpawnStart && c <= g.cfg.SpawnEnd)
			name := "free@2x.png"
			if spawn {
				name = "spawn@2x.png"
			}
			g.drawTile(screen, c, r, name)
		}
	}
}

func (g *Game) drawSettled(screen *ebiten.Image) {
	// settled blocks
	for r := 0; r < g.cfg.NumRows; r++ {
		for c := 0; c < g.cfg.NumColumns; c++ {
			if bt := g.nl.BlockArray[c][r]; bt != nil {
				g.drawTile(screen, c, r, assets.ImageName(bt.Tile))
			}
		}
	}
}

func (g *Game) drawFallingShape(screen *ebiten.Image) {
	// falling shape animation
	if s := g.nl.FallingShape; s != nil {
		for i := range s.Blocks {
			b := &s.Blocks[i]
			if a, ok := g.anims[b.ID]; ok {
				x, y, sc := a.pos()
				g.drawTileAt(screen, x, y, assets.ImageName(b.Tile), sc)
				g.lastVis[b.ID] = [2]float64{x, y}
			} else {
				px, py := g.pointForColumn(b.Column, b.Row)
				g.drawTileAt(screen, px, py, assets.ImageName(b.Tile), 1.0)
				g.lastVis[b.ID] = [2]float64{px, py}
			}
		}
	}
}

func (g *Game) drawGhosts(screen *ebiten.Image) {
	// "explosions"
	for _, gh := range g.ghosts {
		x, y, sc := gh.anim.pos()
		g.drawTileAt(screen, x, y, gh.name, sc)
	}
}

func (g *Game) drawNextShape(screen *ebiten.Image) {
	// preview next shape
	if ns := g.nl.NextShape; ns != nil {
		for i := range ns.Blocks {
			b := &ns.Blocks[i]
			px, py := g.pointForColumn(b.Column, b.Row)
			g.drawTileAt(screen, px, py, assets.ImageName(b.Tile), 1.0)
		}
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	g.drawTimer(screen)

	op := g.tileOp
	op.GeoM.Reset()
	op.GeoM.Translate(8, 8)
	screen.DrawImage(g.panelScore, op)

	if g.nl.Score != g.hudScoreVal {
		g.hudScoreVal = g.nl.Score
		g.hudScoreStr = fmt.Sprintf("Score: %d", g.hudScoreVal)
	}
	if g.nl.Round != g.hudLevelVal {
		g.hudLevelVal = g.nl.Round
		g.hudLevelStr = fmt.Sprintf("Level: %d", g.hudLevelVal)
	}
	g.drawHUDText(screen, 28, g.hudScoreStr, color.White)
	g.drawHUDText(screen, 44, g.hudLevelStr, color.White)

	if g.lb != nil && len(g.lb.Scores) > 0 {
		if g.lbDirty {
			g.lbDirty = false
			mx := min(5, len(g.lb.Scores))
			g.hudLB = g.hudLB[:0]
			for i := 0; i < mx; i++ {
				g.hudLB = append(g.hudLB, fmt.Sprintf("%2d) %4d", i+1, g.lb.Scores[i].Score))
			}
		}

		op := g.tileOp
		op.GeoM.Reset()
		op.GeoM.Translate(8, 52)
		screen.DrawImage(g.panelLB, op)

		g.drawHUDText(screen, 68, "Best:", color.White)
		for i := 0; i < len(g.hudLB); i++ {
			g.drawHUDText(screen, 68+16*(i+1), g.hudLB[i], color.White)
		}
	}
}

var hudFace = text.NewGoXFace(basicfont.Face7x13)

func (g *Game) drawHUDText(dst *ebiten.Image, y int, str string, clr color.Color) {
	op := g.textOp
	op.GeoM.Reset()
	op.GeoM.Translate(16, float64(y)-hudFace.Metrics().HAscent)
	op.ColorScale.Reset()
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(dst, str, hudFace, op)
}

func keyRepeat(key ebiten.Key) bool {
	if inpututil.IsKeyJustPressed(key) {
		return true
	}
	d := inpututil.KeyPressDuration(key)
	if d >= 12 && (d-12)%2 == 0 {
		return true
	}
	return false
}

func (g *Game) drawTile(dst *ebiten.Image, col, row int, name string) {
	img := g.atlas.Get(name)
	op := g.tileOp
	op.GeoM.Reset()
	px, py := g.pointForColumn(col, row)
	op.GeoM.Scale(g.scale, g.scale)
	op.GeoM.Translate(px, py)
	dst.DrawImage(img, op)
}

func (g *Game) drawTileAt(dst *ebiten.Image, px, py float64, name string, scale float64) {
	img := g.atlas.Get(name)
	op := g.tileOp
	op.GeoM.Reset()
	op.GeoM.Scale(g.scale*scale, g.scale*scale)
	op.GeoM.Translate(px, py)
	dst.DrawImage(img, op)
}

func (g *Game) pointForColumn(column, row int) (x, y float64) {
	newRow := row - 1
	newColumn := float64(column) + 0.5
	x = g.layerOffX + newColumn*float64(g.tileSize) + float64(column)
	y = g.layerOffY + float64(newRow)*float64(g.tileSize) + float64(row)
	return x, y
}

// Delegate mapping
func (g *Game) BeginGame() {
	g.nl.BeginGame()
	g.timeLeft = g.tickLength
}

func (g *Game) GameDidBegin(_ *logic.NLine) {
	// field is drawn every frame; nothing to preload here
	// start first shape immediately
	g.NextShape()
}

func (g *Game) NextShape() {
	fs, _ := g.nl.NewShape()
	if fs == nil {
		return
	}
	g.lastVis = make(map[int][2]float64, len(fs.Blocks)) // reset to avoid leak
	g.startSpawnAnim(2, 0.20)
	g.timeLeft = g.tickLength
}

func (g *Game) GameShapeDidMove(_ *logic.NLine) {
	g.startFallingAnim(0.10)
}

func (g *Game) GameShapeDidDrop(_ *logic.NLine) {
	g.sfx.Play("drop.mp3", 1.0)
	g.startFallingAnim(0.05)
	g.settleTimer = 0.05
}

func (g *Game) GameShapeDidLand(_ *logic.NLine) {
	g.nextShapeTimer = 0.01
}

func (g *Game) GameDidEnd(_ *logic.NLine) {
	g.sfx.Play("gameover.mp3", 1.0)
	if g.lb != nil {
		g.lb.Add(g.nl.Score)
		g.lb.Save()
		g.lbDirty = true
	}
	g.BeginGame()
}

func (g *Game) ColorShiftMake(_ *logic.NLine) {}

func (g *Game) GameDidLevelUp(_ *logic.NLine) {
	if g.tickLength >= 2 {
		g.tickLength-- // minus one second per level until we reach 2 sec per shape
	} else {
		g.tickLength -= 0.15
		if g.tickLength < 0.3 {
			g.tickLength = 0.3
		}
	}
	g.sfx.Play("levelup.mp3", 1.0)
}

func (g *Game) LinesMatched(_ *logic.NLine, lines []*logic.Line) {
	g.sfx.Play("bomb.mp3", 1.0)
	for _, ln := range lines {
		for _, t := range ln.Tiles {
			sx, sy := g.pointForColumn(t.Column, t.Row)
			dx := t.Column + g.cfg.RNG.Intn(15) - g.cfg.RNG.Intn(15)
			dy := t.Row + g.cfg.RNG.Intn(15) - g.cfg.RNG.Intn(15)
			ex, ey := g.pointForColumn(dx, dy)
			gh := &Ghost{
				name: assets.ImageName(t.Tile),
				anim: &Anim{
					x0: sx, y0: sy, x1: ex, y1: ey,
					d: 0.20, s0: 1, s1: 5,
				},
			}
			g.ghosts = append(g.ghosts, gh)
		}
	}
	g.nextShapeTimer = 0.40
}

func (g *Game) drawTimer(screen *ebiten.Image) {
	if g.tickLength <= 0 {
		return
	}
	p := 1.0 - (g.timeLeft / g.tickLength)
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}

	vector.StrokeCircle(screen, g.timerPosX, g.timerPosY, g.timerRadius, 6, color.NRGBA{255, 255, 255, 60}, true)

	if p == 0 {
		return
	}

	start := -math.Pi / 2
	end := start + 2*math.Pi*p
	drawArcStroke(screen, g.timerPosX, g.timerPosY, g.timerRadius, float32(6), start, end, color.NRGBA{255, 255, 255, 220})
}

func drawArcStroke(dst *ebiten.Image, cx, cy, r, width float32, start, end float64, col color.Color) {
	if end <= start {
		return
	}
	step := 3.0 * math.Pi / 180.0
	segs := int(math.Ceil((end - start) / step))
	if segs < 1 {
		segs = 1
	}
	a0 := start
	for i := 0; i < segs; i++ {
		a1 := a0 + (end-start)/float64(segs)
		x0 := cx + r*float32(math.Cos(a0))
		y0 := cy + r*float32(math.Sin(a0))
		x1 := cx + r*float32(math.Cos(a1))
		y1 := cy + r*float32(math.Sin(a1))
		vector.StrokeLine(dst, x0, y0, x1, y1, width, col, true)
		a0 = a1
	}
}

func (g *Game) startFallingAnim(duration float64) {
	if s := g.nl.FallingShape; s != nil {
		for i := range s.Blocks {
			b := &s.Blocks[i]
			tx, ty := g.pointForColumn(b.Column, b.Row)
			lv, ok := g.lastVis[b.ID]
			x0, y0 := tx, ty
			if ok {
				x0, y0 = lv[0], lv[1]
			}
			g.anims[b.ID] = &Anim{
				x0: x0, y0: y0, x1: tx, y1: ty,
				d: duration, s0: 1, s1: 1,
			}
		}
	}
}

func (g *Game) startSpawnAnim(offsetRows int, duration float64) {
	if s := g.nl.FallingShape; s != nil {
		for i := range s.Blocks {
			b := &s.Blocks[i]
			tx, ty := g.pointForColumn(b.Column, b.Row)
			sx, sy := g.pointForColumn(b.Column, b.Row-offsetRows)
			g.anims[b.ID] = &Anim{
				x0: sx, y0: sy, x1: tx, y1: ty,
				d: duration, s0: 1, s1: 1,
			}
		}
	}
}
