# 2nLine

Match-3 × tetromino puzzle game on a 10×10 board. Place falling shapes to form lines of 3+ same-coloured tiles — horizontally, vertically, or diagonally — to clear them.

## How to run

Requires Go 1.26+.

```sh
go run ./cmd/2nl
```

Build a standalone binary:

```sh
go build -o 2nl ./cmd/2nl
```

## Controls

| Key | Action |
| --- | --- |
| ← → ↑ ↓ | Move shape |
| `R` | Rotate |
| `Space` | Settle shape |
| `C` | Shift tile colours within the shape |
| `P` | Pause |
| `N` | New game |

## Gameplay

- A shape appears in the spawn zone in the centre of the board. Move, rotate, and settle it before the timer runs out
- Every settled tile stays on the board. Match 3+ tiles of the same colour in a horizontal, vertical, or diagonal line to clear them
- Clearing lines scores `50 × level` points per line. Reaching the level threshold advances the level
- Each level speeds up the timer (down to 0.3 s per shape) and unlocks more shape types — 4 at level 1, up to 11
- When a new shape can't fit in the spawn zone, the game is over. Your top scores are saved locally

## Project layout

```
cmd/2nl/            entry point
internal/logic/     pure game rules (shapes, matching, scoring) — no engine deps
internal/game/      Ebitengine rendering, input, animation
internal/assets/    embedded tile images
internal/sfx/       embedded sound effects
internal/leaderboard/  local high scores (JSON, user config dir)
```

## TODO

Wishlist, not in order of priority:

- [ ] Logo
- [ ] Game / settings selection
- [ ] Icons for pause / new game (reload) / settings / colour shift instead of text boxes
- [ ] Player name input after losing
- [ ] Music
- [ ] Local scoreboard (working one)
- [ ] Gradation of complexity (shapes/colours) — linear or in waves
