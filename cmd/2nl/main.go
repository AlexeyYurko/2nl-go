package main

import (
	"2nline/internal/game"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	g, err := game.New()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowTitle("2nLine (Ebiten)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(g.WindowW(), g.WindowH())
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
