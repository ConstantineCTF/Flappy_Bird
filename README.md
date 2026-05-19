# Flappy Colleï

A Flappy Bird clone built with [Go](https://go.dev) and [Ebitengine v2](https://ebitengine.org/).

Colleï the pink bird flaps through heart-shaped pipes. Features a girly pastel theme with animated wings, pixel-rendered hearts, and a scoring system.

## How to Play

- **SPACE** — Flap / Restart on game over
- Avoid the pipes and the screen edges
- Each pipe passed = 1 point

## Build & Run

```powershell
go build -o flappy.exe .
.\flappy.exe
```

Or simply:

```powershell
go run .
```

## Controls

| Action | Key |
|--------|------|
| Flap | SPACE |
| Restart | SPACE (on game over) |

## Tech Stack

- **Language:** Go 1.24
- **Engine:** [Ebitengine v2.8.8](https://github.com/hajimehoshi/ebiten)
- **Rendering:** All sprites are drawn pixel-by-pixel via `screen.Set()` (no asset files)
