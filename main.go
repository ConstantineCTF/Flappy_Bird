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

const (
	ScreenWidth   = 400
	ScreenHeight  = 600
	Gravity       = 0.45
	FlapVelocity  = -7.5
	MaxUpVelocity = -10.0
	MaxDownVel    = 12.0
	PipeWidth     = 55
	PipeGap       = 160
	PipeSpeed     = 3
	PipeInterval  = 95
	birdX         = 100.0
	birdRadius    = 14.0
	GroundHeight  = 60
	CloudSpeed    = 0.5
	MaxClouds     = 5
	WingFrames    = 3
	HeartSize     = 30
	CapHeight     = 20
)

var (
	SkyColor    = color.RGBA{135, 206, 235, 255}
	SkyBottom   = color.RGBA{200, 230, 255, 255}
	Pink        = color.RGBA{255, 182, 193, 255}
	DarkPink    = color.RGBA{255, 105, 180, 255}
	MintGreen   = color.RGBA{100, 220, 100, 255}
	DarkGreen   = color.RGBA{50, 160, 50, 255}
	Orange      = color.RGBA{255, 165, 0, 255}
	GroundColor = color.RGBA{100, 180, 80, 255}
	GroundDark  = color.RGBA{70, 140, 55, 255}
	GroundBrown = color.RGBA{200, 160, 80, 255}
	PureWhite   = color.RGBA{255, 255, 255, 255}
	PureBlack   = color.RGBA{0, 0, 0, 255}
	Blush       = color.RGBA{255, 150, 150, 180}
	CloudWhite  = color.RGBA{255, 255, 255, 200}
	CapColor    = color.RGBA{40, 140, 40, 255}
)

type Game struct {
	birdY      float64
	birdVY     float64
	pipes      []Pipe
	score      int
	gameOver   bool
	started    bool
	pipeTimer  int
	wingTimer  int
	rng        *rand.Rand

	bgImage      *ebiten.Image
	birdFrames   [WingFrames]*ebiten.Image
	heartImg     *ebiten.Image
	groundImg    *ebiten.Image
	capImg       *ebiten.Image
	cloudImages  []*cloud
	groundScroll float64
}

type Pipe struct {
	x      float64
	gapY   float64
	passed bool
}

type cloud struct {
	x, y  float64
	image *ebiten.Image
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	g := &Game{
		birdY:    ScreenHeight / 3,
		pipes:    []Pipe{},
		rng:      rng,
	}
	g.initAssets(rng)
	return g
}

func (g *Game) initAssets(rng *rand.Rand) {
	g.bgImage = ebiten.NewImage(ScreenWidth, ScreenHeight)
	for y := range ScreenHeight {
		t := float64(y) / float64(ScreenHeight-GroundHeight)
		if t > 1 {
			t = 1
		}
		r := lerp(float64(SkyColor.R), float64(SkyBottom.R), t)
		gg := lerp(float64(SkyColor.G), float64(SkyBottom.G), t)
		b := lerp(float64(SkyColor.B), float64(SkyBottom.B), t)
		clr := color.RGBA{uint8(r), uint8(gg), uint8(b), 255}
		for x := range ScreenWidth {
			g.bgImage.Set(x, y, clr)
		}
	}

	for range MaxClouds {
		cy := 30 + rng.Intn(150)
		c := &cloud{x: float64(rng.Intn(ScreenWidth + 100)), y: float64(cy)}
		c.image = ebiten.NewImage(80, 30)
		drawCloudShape(c.image)
		g.cloudImages = append(g.cloudImages, c)
	}

	for i := range WingFrames {
		img := ebiten.NewImage(36, 36)
		drawBirdOnImage(img, i)
		g.birdFrames[i] = img
	}

	g.heartImg = ebiten.NewImage(HeartSize, HeartSize)
	drawHeartOnImage(g.heartImg, MintGreen)

	g.groundImg = ebiten.NewImage(ScreenWidth, GroundHeight)
	drawGroundOnImage(g.groundImg)

	g.capImg = ebiten.NewImage(PipeWidth+10, CapHeight)
	drawCapOnImage(g.capImg)
}

func drawGroundOnImage(img *ebiten.Image) {
	b := img.Bounds()
	for y := 0; y < b.Dy(); y++ {
		clr := GroundColor
		if y < 4 {
			clr = GroundDark
		} else if y > b.Dy()-8 {
			clr = GroundBrown
		}
		for x := 0; x < b.Dx(); x++ {
			img.Set(x, y, clr)
		}
	}
}

func drawCapOnImage(img *ebiten.Image) {
	b := img.Bounds()
	for y := 0; y < b.Dy(); y++ {
		clr := CapColor
		if y < 3 {
			clr = DarkGreen
		}
		for x := 0; x < b.Dx(); x++ {
			img.Set(x, y, clr)
		}
	}
}

func drawCloudShape(img *ebiten.Image) {
	for _, offset := range [][2]float64{{0, 0}, {-15, -3}, {15, -2}, {-8, -8}, {8, -7}} {
		drawFilledCircle(img, 40+offset[0], 15+offset[1], 10+offset[1]*0.3, CloudWhite)
	}
}

func drawBirdOnImage(img *ebiten.Image, frame int) {
	cx, cy := 18.0, 18.0

	drawFilledCircle(img, cx, cy, birdRadius, Pink)
	drawFilledCircle(img, cx+5, cy-4, 4, PureWhite)
	drawFilledCircle(img, cx+6, cy-4, 2, PureBlack)
	drawTriangleOnImage(img, cx+11, cy-1, cx+16, cy-5, cx+16, cy+3, Orange)
	drawFilledCircle(img, cx-2, cy+5, 3, Blush)

	wingYOff := 0.0
	switch frame {
	case 0:
		wingYOff = -5
	case 1:
		wingYOff = 0
	case 2:
		wingYOff = 4
	}
	drawTriangleOnImage(img, cx-10, cy+wingYOff, cx-18, cy-4+wingYOff, cx-18, cy+4+wingYOff, DarkPink)
}

func drawHeartOnImage(img *ebiten.Image, clr color.Color) {
	for py := range HeartSize {
		for px := range HeartSize {
			cx := float64(px) - float64(HeartSize)/2
			cy := float64(py) - float64(HeartSize)/2
			nx, ny := cx/10.0, cy/9.0
			val := (nx*nx+ny*ny-1)*(nx*nx+ny*ny-1)*(nx*nx+ny*ny-1) - nx*nx*ny*ny*ny
			if val <= 0 {
				img.Set(px, py, clr)
			}
		}
	}
}

func drawFilledCircle(img *ebiten.Image, cx, cy, r float64, clr color.Color) {
	b := img.Bounds()
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				x, y := int(cx+dx), int(cy+dy)
				if x >= 0 && x < b.Dx() && y >= 0 && y < b.Dy() {
					img.Set(x, y, clr)
				}
			}
		}
	}
}

func drawTriangleOnImage(img *ebiten.Image, x1, y1, x2, y2, x3, y3 float64, clr color.Color) {
	b := img.Bounds()
	minX := int(math.Min(math.Min(x1, x2), x3))
	maxX := int(math.Max(math.Max(x1, x2), x3))
	minY := int(math.Min(math.Min(y1, y2), y3))
	maxY := int(math.Max(math.Max(y1, y2), y3))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if pointInTriangle(float64(x), float64(y), x1, y1, x2, y2, x3, y3) {
				if x >= 0 && x < b.Dx() && y >= 0 && y < b.Dy() {
					img.Set(x, y, clr)
				}
			}
		}
	}
}

func pointInTriangle(px, py, x1, y1, x2, y2, x3, y3 float64) bool {
	area := 0.5 * (-y2*x3 + y1*(-x2+x3) + x1*(y2-y3) + x2*y3 - x3*y1)
	if area == 0 {
		return false
	}
	s := 1 / (2 * area) * (y1*x3 - x1*y3 + (y3-y1)*px + (x1-x3)*py)
	t := 1 / (2 * area) * (x1*y2 - y1*x2 + (y1-y2)*px + (x2-x1)*py)
	return s >= 0 && t >= 0 && (s+t) <= 1
}

func (g *Game) Update() error {
	if !g.started {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.started = true
		}
		return nil
	}

	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			*g = *NewGame()
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.birdVY = FlapVelocity
	}

	g.birdVY += Gravity
	if g.birdVY > MaxDownVel {
		g.birdVY = MaxDownVel
	}
	g.birdY += g.birdVY
	g.wingTimer++

	for _, c := range g.cloudImages {
		c.x -= CloudSpeed
		if c.x < -100 {
			c.x = ScreenWidth + float64(g.rng.Intn(100))
			c.y = 30 + float64(g.rng.Intn(150))
		}
	}

	g.groundScroll -= PipeSpeed
	if g.groundScroll <= -ScreenWidth {
		g.groundScroll += ScreenWidth
	}

	g.pipeTimer++
	if g.pipeTimer > PipeInterval {
		minGap := 80.0
		maxGap := float64(ScreenHeight - GroundHeight - PipeGap - 80)
		if maxGap < minGap {
			maxGap = minGap
		}
		gapY := minGap + float64(g.rng.Intn(int(maxGap-minGap+1)))
		g.pipes = append(g.pipes, Pipe{x: ScreenWidth, gapY: gapY, passed: false})
		g.pipeTimer = 0
	}

	for i := range g.pipes {
		g.pipes[i].x -= PipeSpeed
		if !g.pipes[i].passed && g.pipes[i].x+PipeWidth/2 < birdX {
			g.pipes[i].passed = true
			g.score++
		}
	}

	if len(g.pipes) > 0 && g.pipes[0].x < -PipeWidth {
		g.pipes = g.pipes[1:]
	}

	g.checkCollisions()
	return nil
}

func (g *Game) checkCollisions() {
	groundTop := ScreenHeight - GroundHeight
	if g.birdY+birdRadius >= float64(groundTop) || g.birdY-birdRadius <= 0 {
		g.gameOver = true
		return
	}

	for _, p := range g.pipes {
		if birdX+birdRadius > p.x && birdX-birdRadius < p.x+PipeWidth {
			if g.birdY-birdRadius < p.gapY || g.birdY+birdRadius > p.gapY+PipeGap {
				g.gameOver = true
				return
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(g.bgImage, op)

	for _, c := range g.cloudImages {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(c.x, c.y)
		screen.DrawImage(c.image, op)
	}

	groundY := float64(ScreenHeight - GroundHeight)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.groundScroll, groundY)
	screen.DrawImage(g.groundImg, op)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.groundScroll+ScreenWidth, groundY)
	screen.DrawImage(g.groundImg, op)

	halfH := float64(HeartSize) / 2
	for _, p := range g.pipes {
		for y := 0.0; y < p.gapY-HeartSize; y += HeartSize {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.x+PipeWidth/2-halfH, y)
			screen.DrawImage(g.heartImg, op)
		}
		for y := p.gapY + PipeGap; y < float64(ScreenHeight-GroundHeight); y += HeartSize {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.x+PipeWidth/2-halfH, y)
			screen.DrawImage(g.heartImg, op)
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x-5, p.gapY-CapHeight)
		screen.DrawImage(g.capImg, op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x-5, p.gapY+PipeGap)
		screen.DrawImage(g.capImg, op)
	}

	frame := 1
	if g.started && !g.gameOver {
		frame = (g.wingTimer / 6) % WingFrames
	}

	angle := g.birdVY * 3
	if angle < -30 {
		angle = -30
	}
	if angle > 60 {
		angle = 60
	}

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-18, -18)
	op.GeoM.Rotate(angle * math.Pi / 180)
	op.GeoM.Translate(birdX, g.birdY)
	screen.DrawImage(g.birdFrames[frame], op)

	if !g.started {
		ebitenutil.DebugPrintAt(screen, "FLAPPY COLLEI", ScreenWidth/2-55, ScreenHeight/3)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to start", ScreenWidth/2-75, ScreenHeight/3+25)
	}

	scoreStr := fmt.Sprintf("SCORE: %d", g.score)
	scoreW := len(scoreStr) * 6
	ebitenutil.DebugPrintAt(screen, scoreStr, ScreenWidth/2-scoreW/2, 20)

	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "GAME OVER", ScreenWidth/2-40, ScreenHeight/2-10)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to restart", ScreenWidth/2-80, ScreenHeight/2+15)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Flappy Colleï")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
