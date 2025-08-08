package sfx

import (
	"bytes"
	"embed"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/go-mp3"
)

//go:embed audio/*.mp3
var sfxFS embed.FS

type SFX struct {
	ctx      *audio.Context
	data     map[string][]byte
	active   []*audio.Player
	mu       sync.Mutex
	sampleHz int
}

func LoadSFX() *SFX {
	entries, err := sfxFS.ReadDir("audio")
	if err != nil {
		log.Fatalf("read audio dir: %v", err)
	}

	s := &SFX{data: map[string][]byte{}}
	for i, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".mp3") {
			continue
		}
		b, err := sfxFS.ReadFile("audio/" + e.Name())
		if err != nil {
			log.Fatalf("missing sfx %s: %v", e.Name(), err)
		}
		dec, err := mp3.NewDecoder(bytes.NewReader(b))
		if err != nil {
			log.Fatalf("decode mp3 %s: %v", e.Name(), err)
		}
		if i == 0 {
			s.sampleHz = dec.SampleRate()
			s.ctx = audio.NewContext(s.sampleHz)
		} else if dec.SampleRate() != s.sampleHz {
			log.Fatalf("mp3 %s sample rate mismatch: %d != %d", e.Name(), dec.SampleRate(), s.sampleHz)
		}
		pcm, err := io.ReadAll(dec)
		if err != nil {
			log.Fatalf("read mp3 %s: %v", e.Name(), err)
		}
		s.data[e.Name()] = pcm
	}
	return s
}

func (s *SFX) Play(name string, vol float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.data[name]
	if !ok {
		return
	}
	p := s.ctx.NewPlayerFromBytes(data)
	if vol <= 0 {
		vol = 1.0
	}
	p.SetVolume(vol)
	p.Play()
	s.active = append(s.active, p)
}

func (s *SFX) Update() {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := 0
	for _, p := range s.active {
		if p.IsPlaying() {
			s.active[j] = p
			j++
		} else {
			_ = p.Close()
		}
	}
	s.active = s.active[:j]
}

func (s *SFX) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.active {
		_ = p.Close()
	}
	s.active = nil
}
