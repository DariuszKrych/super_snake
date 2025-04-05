package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"snake-game/internal/game"
)

const (
	GridCellSize = 20 // Visual size of each grid cell in pixels
)

var (
	bgColor           = color.RGBA{R: 15, G: 15, B: 25, A: 255}    // Dark blue-ish background
	gridColor         = color.RGBA{R: 50, G: 50, B: 70, A: 255}    // Faint grid lines
	wallColor         = color.RGBA{R: 100, G: 100, B: 120, A: 255} // Color for boundaries
	playerHeadColor   = color.RGBA{R: 0, G: 200, B: 50, A: 255}
	playerBodyColor   = color.RGBA{R: 0, G: 255, B: 80, A: 255}
	enemyHeadColor    = color.RGBA{R: 200, G: 50, B: 0, A: 255}  // Example enemy color
	enemyBodyColor    = color.RGBA{R: 255, G: 80, B: 0, A: 255}  // Example enemy color
	foodStandardColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}   // Red
	foodSpeedColor    = color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange
	foodSlowColor     = color.RGBA{R: 0, G: 191, B: 255, A: 255} // Deep Sky Blue
)

// DrawGame renders the entire game state.
func DrawGame(screen *ebiten.Image, state game.RenderableState) {
	// screenWidth, screenHeight := screen.Size() // Remove this line

	// 1. Draw Background
	screen.Fill(bgColor)

	// 2. Draw Grid (Optional, can be subtle)
	// drawGrid(screen, state.GridWidth, state.GridHeight, screenWidth, screenHeight)

	// 3. Draw Walls/Boundaries
	drawWalls(screen, state.GridWidth, state.GridHeight)

	// 4. Draw Food (Iterate over slice)
	// if state.Food != nil { // Old check
	// 	drawFood(screen, *state.Food)
	// }
	for _, food := range state.FoodItems {
		if food != nil { // Check if pointer is valid
			drawFood(screen, *food) // Dereference pointer to pass game.Food
		}
	}

	// 5. Draw Enemy Snakes
	for _, enemy := range state.EnemySnakes {
		if enemy != nil {
			drawSnake(screen, *enemy, enemyHeadColor, enemyBodyColor)
		}
	}

	// 6. Draw Player Snake (drawn last to be on top)
	if state.PlayerSnake != nil {
		drawSnake(screen, *state.PlayerSnake, playerHeadColor, playerBodyColor)
	}

	// 7. Draw HUD (Score, etc.) - To be implemented later
	// drawHUD(screen, state.Score, state.PlayerSpeedFactor, state.SpeedEffectDuration)
}

// drawGrid draws faint grid lines (optional visual aid)
func drawGrid(screen *ebiten.Image, gridW, gridH, screenW, screenH int) {
	// Vertical lines
	for x := 0; x <= gridW; x++ {
		fx := float32(x * GridCellSize)
		vector.StrokeLine(screen, fx, 0, fx, float32(screenH), 1, gridColor, false)
	}
	// Horizontal lines
	for y := 0; y <= gridH; y++ {
		fy := float32(y * GridCellSize)
		vector.StrokeLine(screen, 0, fy, float32(screenW), fy, 1, gridColor, false)
	}
}

// drawWalls draws the boundaries of the game area.
func drawWalls(screen *ebiten.Image, gridW, gridH int) {
	thickness := float32(2) // Wall thickness
	w := float32(gridW * GridCellSize)
	h := float32(gridH * GridCellSize)

	// Top
	vector.DrawFilledRect(screen, 0, 0, w, thickness, wallColor, false)
	// Bottom
	vector.DrawFilledRect(screen, 0, h-thickness, w, thickness, wallColor, false)
	// Left
	vector.DrawFilledRect(screen, 0, 0, thickness, h, wallColor, false)
	// Right
	vector.DrawFilledRect(screen, w-thickness, 0, thickness, h, wallColor, false)
}

// drawSnake draws a single snake.
func drawSnake(screen *ebiten.Image, s game.Snake, headClr, bodyClr color.Color) {
	if len(s.Body) == 0 {
		return
	}

	// Draw body segments first
	for i := 1; i < len(s.Body); i++ {
		segment := s.Body[i]
		vector.DrawFilledRect(
			screen,
			float32(segment.X*GridCellSize)+1, // +1, -2 for slight padding
			float32(segment.Y*GridCellSize)+1,
			float32(GridCellSize-2),
			float32(GridCellSize-2),
			bodyClr,
			false,
		)
	}

	// Draw head last (on top)
	head := s.Body[0]
	vector.DrawFilledRect(
		screen,
		float32(head.X*GridCellSize)+1,
		float32(head.Y*GridCellSize)+1,
		float32(GridCellSize-2),
		float32(GridCellSize-2),
		headClr,
		false,
	)
}

// drawFood draws a food item.
func drawFood(screen *ebiten.Image, f game.Food) {
	var clr color.Color
	switch f.Type {
	case game.FoodTypeStandard:
		clr = foodStandardColor
	case game.FoodTypeSpeedUp:
		clr = foodSpeedColor
	case game.FoodTypeSlowDown:
		clr = foodSlowColor
	default:
		clr = color.White // Fallback
	}

	vector.DrawFilledRect(
		screen,
		float32(f.Pos.X*GridCellSize)+2, // +2, -4 for slightly smaller food
		float32(f.Pos.Y*GridCellSize)+2,
		float32(GridCellSize-4),
		float32(GridCellSize-4),
		clr,
		false,
	)
}

// TODO: drawHUD function
// func drawHUD(screen *ebiten.Image, score int, speedFactor float64, effectDuration time.Duration) { ... }
