package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"snake-game/internal/assets"
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

// DrawGame renders the entire game state using assets.
func DrawGame(screen *ebiten.Image, state game.RenderableState, assets *assets.Manager) {
	// screenWidth, screenHeight := screen.Size() // Remove this line

	// 1. Draw Background
	if assets.Background != nil {
		// Basic tiling or stretching - adjust as needed
		bgWidth, bgHeight := assets.Background.Size()
		screenWidth, screenHeight := screen.Size()
		// op := &ebiten.DrawImageOptions{} // Remove this unused declaration
		// Simple stretch example:
		// op.GeoM.Scale(float64(screenWidth)/float64(bgWidth), float64(screenHeight)/float64(bgHeight))
		// Tiling example:
		maxX := screenWidth / bgWidth
		maxY := screenHeight / bgHeight
		for y := 0; y <= maxY; y++ {
			for x := 0; x <= maxX; x++ {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x*bgWidth), float64(y*bgHeight))
				screen.DrawImage(assets.Background, op)
			}
		}
	} else {
		screen.Fill(bgColor) // Fallback background color
	}

	// 2. Draw Grid (Optional, can be subtle)
	// drawGrid(screen, state.GridWidth, state.GridHeight, screenWidth, screenHeight)

	// 3. Draw Walls/Boundaries
	drawWalls(screen, state.GridWidth, state.GridHeight, assets)

	// 4. Draw Food (Iterate over slice)
	// if state.Food != nil { // Old check
	// 	drawFood(screen, *state.Food)
	// }
	for _, food := range state.FoodItems {
		if food != nil { // Check if pointer is valid
			drawFood(screen, *food, assets) // Dereference pointer to pass game.Food
		}
	}

	// 5. Draw Enemy Snakes
	for _, enemy := range state.EnemySnakes {
		if enemy != nil {
			drawSnake(screen, *enemy, assets)
		}
	}

	// 6. Draw Player Snake (drawn last to be on top)
	if state.PlayerSnake != nil {
		drawSnake(screen, *state.PlayerSnake, assets)
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
func drawWalls(screen *ebiten.Image, gridW, gridH int, assets *assets.Manager) {
	// Use wall sprite if available, otherwise fallback to colored rects
	if assets.Wall != nil {
		// TODO: Implement drawing walls using the assets.Wall sprite
		// This might involve drawing tiles or stretching the sprite.
		// For now, fallback to simple rects.
		drawWallRects(screen, gridW, gridH)
	} else {
		drawWallRects(screen, gridW, gridH)
	}
}

// drawWallRects draws simple rectangles for walls (fallback).
func drawWallRects(screen *ebiten.Image, gridW, gridH int) {
	thickness := float32(2)
	w := float32(gridW * GridCellSize)
	h := float32(gridH * GridCellSize)
	vector.DrawFilledRect(screen, 0, 0, w, thickness, wallColor, false)
	vector.DrawFilledRect(screen, 0, h-thickness, w, thickness, wallColor, false)
	vector.DrawFilledRect(screen, 0, 0, thickness, h, wallColor, false)
	vector.DrawFilledRect(screen, w-thickness, 0, thickness, h, wallColor, false)
}

// drawSnake draws a single snake using sprites.
func drawSnake(screen *ebiten.Image, s game.Snake, assets *assets.Manager) {
	if len(s.Body) == 0 || assets.SnakeBody == nil || assets.SnakeHead == nil {
		return // Cannot draw without assets or body
	}

	bodyW, bodyH := assets.SnakeBody.Size()
	headW, headH := assets.SnakeHead.Size()

	// --- Draw Body ---
	for i := 1; i < len(s.Body); i++ {
		segment := s.Body[i]
		prevSegment := s.Body[i-1] // Get the segment in front of this one

		op := &ebiten.DrawImageOptions{}
		// Center the sprite within the grid cell
		tx := float64(segment.X*GridCellSize) + float64(GridCellSize-bodyW)/2.0
		ty := float64(segment.Y*GridCellSize) + float64(GridCellSize-bodyH)/2.0

		// --- Determine Body Rotation ---
		var bodyAngle float64 = 0
		if prevSegment.X == segment.X { // Moving vertically (previous is above or below)
			bodyAngle = math.Pi / 2 // Rotate 90 degrees
		} // Else: Moving horizontally, angle remains 0 (assuming horizontal sprite)

		// Apply rotation if needed
		if bodyAngle != 0 {
			centerX := float64(bodyW) / 2.0
			centerY := float64(bodyH) / 2.0
			op.GeoM.Translate(-centerX, -centerY)
			op.GeoM.Rotate(bodyAngle)
			op.GeoM.Translate(centerX, centerY)
		}

		// Apply translation
		op.GeoM.Translate(tx, ty)
		// TODO: Add rotation for body segments based on previous segment? (Advanced) // Already doing basic rotation
		screen.DrawImage(assets.SnakeBody, op)
	}

	// --- Draw Head ---
	head := s.Body[0]
	op := &ebiten.DrawImageOptions{}
	// Calculate translation to center the head sprite
	tx := float64(head.X*GridCellSize) + float64(GridCellSize-headW)/2.0
	ty := float64(head.Y*GridCellSize) + float64(GridCellSize-headH)/2.0

	// Calculate rotation based on direction
	var angle float64
	switch s.Direction {
	case game.DirUp:
		angle = -math.Pi / 2 // -90 degrees
	case game.DirDown:
		angle = math.Pi / 2 // 90 degrees
	case game.DirLeft:
		angle = math.Pi // 180 degrees
	case game.DirRight:
		angle = 0 // Assume head sprite faces right by default
	default:
		angle = 0 // Or use previous direction?
	}

	// Apply rotation around the center of the sprite
	centerX := float64(headW) / 2.0
	centerY := float64(headH) / 2.0
	op.GeoM.Translate(-centerX, -centerY) // Move rotation center to origin
	op.GeoM.Rotate(angle)                 // Rotate
	op.GeoM.Translate(centerX, centerY)   // Move back

	// Apply translation to position the head on the grid
	op.GeoM.Translate(tx, ty)

	screen.DrawImage(assets.SnakeHead, op)
}

// drawFood draws a food item using sprites.
func drawFood(screen *ebiten.Image, f game.Food, assets *assets.Manager) {
	var img *ebiten.Image
	switch f.Type {
	case game.FoodTypeStandard:
		img = assets.FoodStandard
	case game.FoodTypeSpeedUp:
		img = assets.FoodSpeedUp
	case game.FoodTypeSlowDown:
		img = assets.FoodSlowDown
	default:
		return // Don't draw unknown food types
	}

	if img == nil {
		return // Don't draw if asset is missing
	}

	imgW, imgH := img.Size()
	op := &ebiten.DrawImageOptions{}
	// Center the sprite
	tx := float64(f.Pos.X*GridCellSize) + float64(GridCellSize-imgW)/2.0
	ty := float64(f.Pos.Y*GridCellSize) + float64(GridCellSize-imgH)/2.0
	op.GeoM.Translate(tx, ty)

	screen.DrawImage(img, op)
}

// TODO: drawHUD function
// func drawHUD(screen *ebiten.Image, score int, speedFactor float64, effectDuration time.Duration) { ... }
