package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ScreenWidth    = 1920 // Adjust if needed, will run fullscreen
	ScreenHeight   = 1080 // Adjust if needed, will run fullscreen
	GridSize       = 20   // Size of each grid cell (and snake segment/food)
	InitialSpeed   = 8    // Ticks per second (initial snake movement speed)
	SpeedIncrement = 0.5  // How much speed increases per food eaten
	MaxSpeed       = 20   // Maximum speed
	InitialSnakeLen = 3
)

// Calculate grid dimensions based on screen size and grid size
var (
	GridWidth  = ScreenWidth / GridSize
	GridHeight = ScreenHeight / GridSize
)

// Direction constants
type Direction int

const (
	DirNone Direction = iota
	DirUp
	DirDown
	DirLeft
	DirRight
)

// Position represents a point on the grid
type Position struct {
	X, Y int
}

// Snake struct
type Snake struct {
	Body      []Position
	Direction Direction
	NextDir   Direction // Buffer for next direction input
}

// Food struct
type Food struct {
	Pos Position
}

// Game struct holds the entire game state
type Game struct {
	Snake        *Snake
	Food         *Food
	Score        int
	Speed        float64 // Ticks per second
	MoveTicker   *time.Ticker
	GameOver     bool
	ScreenWidth  int
	ScreenHeight int
	needsReset   bool // Flag to signal game reset
}

// NewGame initializes a new game state
func NewGame() *Game {
	g := &Game{
		ScreenWidth:  ScreenWidth,
		ScreenHeight: ScreenHeight,
		needsReset:   true, // Start with initial setup
	}
	g.Reset() // Initialize game elements
	return g
}

// Reset initializes or resets the game state
func (g *Game) Reset() {
	// Initialize snake
	startX, startY := GridWidth/2, GridHeight/2
	initialBody := make([]Position, 0, InitialSnakeLen)
	for i := 0; i < InitialSnakeLen; i++ {
		initialBody = append(initialBody, Position{X: startX - i, Y: startY})
	}
	g.Snake = &Snake{
		Body:      initialBody,
		Direction: DirRight,
		NextDir:   DirRight,
	}

	g.Food = &Food{}
	g.spawnFood() // Initial food placement

	g.Score = 0
	g.Speed = InitialSpeed
	if g.MoveTicker != nil {
		g.MoveTicker.Stop()
	}
	g.MoveTicker = time.NewTicker(time.Second / time.Duration(g.Speed))
	g.GameOver = false
	g.needsReset = false
}

// spawnFood places the food randomly, avoiding the snake
func (g *Game) spawnFood() {
	// Seed random number generator once at the start in main
	// rand.Seed(time.Now().UnixNano()) // Avoid re-seeding frequently
	occupied := make(map[Position]bool)
	for _, segment := range g.Snake.Body {
		occupied[segment] = true
	}

	for {
		newPos := Position{
			X: rand.Intn(GridWidth),
			Y: rand.Intn(GridHeight),
		}
		if !occupied[newPos] {
			g.Food.Pos = newPos
			break
		}
	}
}

// updateSpeed adjusts the game speed based on score
func (g *Game) updateSpeed() {
	newSpeed := InitialSpeed + float64(g.Score)*SpeedIncrement
	if newSpeed > MaxSpeed {
		newSpeed = MaxSpeed
	}
	if newSpeed != g.Speed {
		g.Speed = newSpeed
		g.MoveTicker.Reset(time.Second / time.Duration(g.Speed))
	}
}

// HandleInput processes user key presses
func (g *Game) HandleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && g.Snake.Direction != DirDown {
		g.Snake.NextDir = DirUp
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && g.Snake.Direction != DirUp {
		g.Snake.NextDir = DirDown
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && g.Snake.Direction != DirRight {
		g.Snake.NextDir = DirLeft
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && g.Snake.Direction != DirLeft {
		g.Snake.NextDir = DirRight
	}

	// Allow restarting the game after game over
	if g.GameOver && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.needsReset = true
	}
}

// Update game logic
func (g *Game) Update() error {
	if g.needsReset {
		g.Reset()
	}

	g.HandleInput()

	if g.GameOver {
		return nil // Stop updates if game over
	}

	// Check the ticker to see if it's time to move
	select {
	case <-g.MoveTicker.C:
		// Update direction based on buffered input
		g.Snake.Direction = g.Snake.NextDir

		// Calculate new head position
		head := g.Snake.Body[0]
		newHead := head
		switch g.Snake.Direction {
		case DirUp:
			newHead.Y--
		case DirDown:
			newHead.Y++
		case DirLeft:
			newHead.X--
		case DirRight:
			newHead.X++
		}

		// Check boundary collision
		if newHead.X < 0 || newHead.X >= GridWidth || newHead.Y < 0 || newHead.Y >= GridHeight {
			g.GameOver = true
			if g.MoveTicker != nil { // Ensure ticker exists before stopping
        		g.MoveTicker.Stop()
    		}
			return nil
		}

		// Check self collision
		for i := 1; i < len(g.Snake.Body); i++ {
			if newHead == g.Snake.Body[i] {
				g.GameOver = true
				if g.MoveTicker != nil { // Ensure ticker exists before stopping
        			g.MoveTicker.Stop()
    			}
				return nil
			}
		}

		// Check food collision
		ateFood := false
		if newHead == g.Food.Pos {
			ateFood = true
			g.Score++
			g.spawnFood()
			g.updateSpeed()
		}

		// Update snake body
		// Prepend new head
		g.Snake.Body = append([]Position{newHead}, g.Snake.Body...)

		// Remove tail if food was not eaten
		if !ateFood {
			g.Snake.Body = g.Snake.Body[:len(g.Snake.Body)-1]
		}

	default:
		// Not time to move yet
	}

	return nil
}

// Draw renders the game state
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black) // Background

	// Draw snake
	headColor := color.RGBA{R: 0, G: 180, B: 0, A: 255} // Darker green for head
	bodyColor := color.RGBA{R: 0, G: 255, B: 0, A: 255} // Bright green for body
	for i, segment := range g.Snake.Body {
		clr := bodyColor
		if i == 0 {
			clr = headColor
		}
		// Draw slightly smaller squares for visual separation
		vector.DrawFilledRect(screen, float32(segment.X*GridSize)+1, float32(segment.Y*GridSize)+1, float32(GridSize-2), float32(GridSize-2), clr, false)
	}

	// Draw food
	foodColor := color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red
	vector.DrawFilledRect(screen, float32(g.Food.Pos.X*GridSize)+1, float32(g.Food.Pos.Y*GridSize)+1, float32(GridSize-2), float32(GridSize-2), foodColor, false)

	// Draw score
	scoreStr := fmt.Sprintf("Score: %d", g.Score)
	ebitenutil.DebugPrintAt(screen, scoreStr, 10, 10)

	// Draw Game Over message
	if g.GameOver {
		// Use ebitenutil for simple centered text (may need adjusting for exact centering)
        gameOverTitle := "Game Over!"
        finalScore := fmt.Sprintf("Final Score: %d", g.Score)
        restartMsg := "Press SPACE to Restart"

        titleWidth := len(gameOverTitle) * 8 // Approx using DebugPrint font size
        scoreWidth := len(finalScore) * 8
        restartWidth := len(restartMsg) * 8

        titleX := (ScreenWidth - titleWidth) / 2
        scoreX := (ScreenWidth - scoreWidth) / 2
        restartX := (ScreenWidth - restartWidth) / 2

        titleY := ScreenHeight/2 - 20
        scoreY := ScreenHeight / 2
        restartY := ScreenHeight/2 + 20


		ebitenutil.DebugPrintAt(screen, gameOverTitle, titleX, titleY)
        ebitenutil.DebugPrintAt(screen, finalScore, scoreX, scoreY)
        ebitenutil.DebugPrintAt(screen, restartMsg, restartX, restartY)
	}
}

// Layout defines the logical screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    // This ensures the game renders consistently regardless of window size changes
    // if not in fullscreen, or if fullscreen resolution differs slightly.
	return ScreenWidth, ScreenHeight
}

func main() {
    // Seed random number generator once at the start
    rand.Seed(time.Now().UnixNano())

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Super Snake")
	ebiten.SetFullscreen(true) // Set to fullscreen mode
    // On Linux, Ebiten tries to find the best fullscreen mode.
    // If ScreenWidth/Height don't match an available mode, it might pick the closest.

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
} 