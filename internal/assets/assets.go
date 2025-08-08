package assets

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed images/*.png
var assetsFS embed.FS

type Atlas struct {
	img map[string]*ebiten.Image
}

func LoadAtlas() *Atlas {
	entries, err := assetsFS.ReadDir("images")
	if err != nil {
		log.Fatal(err)
	}
	m := map[string]*ebiten.Image{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".png") {
			continue
		}
		b, err := assetsFS.ReadFile("images/" + e.Name())
		if err != nil {
			log.Fatalf("missing asset %s: %v", e.Name(), err)
		}
		img, _, err := image.Decode(bytes.NewReader(b))
		if err != nil {
			log.Fatalf("decode %s: %v", e.Name(), err)
		}
		m[e.Name()] = ebiten.NewImageFromImage(img)
	}
	return &Atlas{img: m}
}

func (a *Atlas) Get(name string) *ebiten.Image {
	im := a.img[name]
	if im == nil {
		im = a.img["blue@2x.png"]
	}
	return im
}

var tileToName = [...]string{
	"", "blue@2x.png", "green@2x.png", "purple@2x.png",
	"red@2x.png", "white@2x.png", "yellow@2x.png",
}

func ImageName(tile int) string {
	if tile <= 0 || tile >= len(tileToName) {
		return tileToName[1]
	}
	return tileToName[tile]
}
