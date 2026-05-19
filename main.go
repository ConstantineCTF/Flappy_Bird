package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Game constants
const (
	ScreenWidth  = 400
	ScreenHeight = 600
	Gravity      = 0.5
	FlapVelocity = -8
	PipeWidth    = 60
	PipeGap      = 150
	PipeSpeed    = 2
	PipeInterval = 100
	birdX        = 100.0
	birdRadius   = 15.0
)

// Colors for girly theme
var (
	Pink      = color.RGBA{255, 182, 193, 255} // Light pink for Colleï
	MintGreen = color.RGBA{152, 255, 152, 255} // Pipes
)

// Game struct holds game state
type Game struct {
	birdY      float64
	birdVY     float64
	pipes      []Pipe
	score      int
	gameOver   bool
	pipeTimer  int
	wingFrame  int
	randSource *rand.Rand
}

// Pipe struct for obstacles
type Pipe struct {
	x      float64
	gapY   float64
	passed bool
}

// NewGame initializes the game
func NewGame() *Game {
	return &Game{
		birdY:      ScreenHeight / 2,
		birdVY:     0,
		pipes:      []Pipe{},
		score:      0,
		gameOver:   false,
		pipeTimer:  0,
		wingFrame:  0,
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Update updates the game state
func (g *Game) Update() error {
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			// Restart game
			*g = *NewGame()
		}
		return nil
	}

	// Flap on spacebar
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.birdVY = FlapVelocity
	}

	// Update bird
	g.birdVY += Gravity
	g.birdY += g.birdVY
	g.wingFrame = (g.wingFrame + 1) % 20 // Wing animation

	// Generate pipes
	g.pipeTimer++
	if g.pipeTimer > PipeInterval {
		gapY := float64(100 + g.randSource.Intn(ScreenHeight-200-PipeGap))
		g.pipes = append(g.pipes, Pipe{x: ScreenWidth, gapY: gapY, passed: false})
		g.pipeTimer = 0
	}

	// Update pipes
	for i := range g.pipes {
		g.pipes[i].x -= PipeSpeed
		// Score when passing pipe
		if !g.pipes[i].passed && g.pipes[i].x+PipeWidth/2 < 100 {
			g.pipes[i].passed = true
			g.score++
		}
	}

	// Remove off-screen pipes
	if len(g.pipes) > 0 && g.pipes[0].x < -PipeWidth {
		g.pipes = g.pipes[1:]
	}

	// Collision detection
	for _, p := range g.pipes {
		pipeX := p.x
		gapY := p.gapY
		// Check collision with top and bottom pipes
		if birdX+birdRadius > pipeX && birdX-birdRadius < pipeX+PipeWidth {
			if g.birdY-birdRadius < gapY || g.birdY+birdRadius > gapY+PipeGap {
				g.gameOver = true
			}
		}
	}

	// Check out of bounds with exact visual contact
	if g.birdY-birdRadius <= 0 || g.birdY+birdRadius >= ScreenHeight {
		g.gameOver = true
	}

	return nil
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Draw background
	screen.Fill(color.White)

	// Draw pipes (heart-shaped)
	for _, p := range g.pipes {
		// Top pipe
		for y := 0.0; y < p.gapY; y += 10 {
			g.drawHeart(screen, p.x+PipeWidth/2, y, MintGreen)
		}
		// Bottom pipe
		for y := p.gapY + PipeGap; y < ScreenHeight; y += 10 {
			g.drawHeart(screen, p.x+PipeWidth/2, y, MintGreen)
		}
	}

	// Draw Colleï (pink bird with wings)
	g.drawBird(screen, 100, g.birdY)

	// Draw score
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.score), 10, 10)

	// Draw game over screen
	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "Game Over! Press SPACE to restart", ScreenWidth/2-100, ScreenHeight/2)
	}
}

// drawBird draws Colleï with animated wings
func (g *Game) drawBird(screen *ebiten.Image, x, y float64) {
	// Body (circle)
	g.drawCircle(screen, x, y, 15, Pink)

	// Eye
	g.drawCircle(screen, x+5, y-5, 3, color.White)
	g.drawCircle(screen, x+5, y-5, 1, color.Black)

	// Beak
	g.drawTriangle(screen, x+10, y, x+15, y-3, x+15, y+3, color.RGBA{255, 165, 0, 255})

	// Wings (animated)
	wingOffset := 5.0
	if g.wingFrame < 10 {
		wingOffset = -5.0
	}
	g.drawWing(screen, x-10, y+wingOffset, Pink)
}

// drawCircle draws a filled circle
func (g *Game) drawCircle(screen *ebiten.Image, cx, cy, r float64, clr color.Color) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				screen.Set(int(cx+x), int(cy+y), clr)
			}
		}
	}
}

// drawTriangle draws a filled triangle
func (g *Game) drawTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float64, clr color.Color) {
	minX := math.Min(math.Min(x1, x2), x3)
	maxX := math.Max(math.Max(x1, x2), x3)
	minY := math.Min(math.Min(y1, y2), y3)
	maxY := math.Max(math.Max(y1, y2), y3)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if g.pointInTriangle(x, y, x1, y1, x2, y2, x3, y3) {
				screen.Set(int(x), int(y), clr)
			}
		}
	}
}

// pointInTriangle checks if a point is inside a triangle
func (g *Game) pointInTriangle(px, py, x1, y1, x2, y2, x3, y3 float64) bool {
	area := 0.5 * (-y2*x3 + y1*(-x2+x3) + x1*(y2-y3) + x2*y3 - x3*y1)
	if area == 0 {
		return false
	}
	s := 1 / (2 * area) * (y1*x3 - x1*y3 + (y3-y1)*px + (x1-x3)*py)
	t := 1 / (2 * area) * (x1*y2 - y1*x2 + (y1-y2)*px + (x2-x1)*py)
	return s >= 0 && t >= 0 && (s+t) <= 1
}

// drawWing draws a wing
func (g *Game) drawWing(screen *ebiten.Image, x, y float64, clr color.Color) {
	g.drawTriangle(screen, x, y, x-10, y-5, x-10, y+5, clr)
}

// drawHeart draws a heart shape
func (g *Game) drawHeart(screen *ebiten.Image, cx, cy float64, clr color.Color) {
	for y := -10.0; y <= 10; y++ {
		for x := -10.0; x <= 10; x++ {
			if (x*x+y*y-100)*(x*x+y*y-100)*(x*x+y*y-100) < x*x*y*y*y {
				screen.Set(int(cx+x), int(cy+y), clr)
			}
		}
	}
}

// Layout sets the screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Flappy Colleï")
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
