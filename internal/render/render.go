package render

import (
	"fmt"
	"image/color"
	"math"
	"time" // Import time package

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"snake-game/internal/assets"
	"snake-game/internal/game"
)

const (
	GridCellSize = 20 // Visual size of each grid cell in pixels
)

var (
	bgColor            = color.RGBA{R: 15, G: 15, B: 25, A: 255}    // Dark blue-ish background
	gridColor          = color.RGBA{R: 50, G: 50, B: 70, A: 255}    // Faint grid lines
	wallColor          = color.RGBA{R: 100, G: 100, B: 120, A: 255} // Color for boundaries
	playerHeadColor    = color.RGBA{R: 0, G: 200, B: 50, A: 255}
	playerBodyColor    = color.RGBA{R: 0, G: 255, B: 80, A: 255}
	enemyHeadColor     = color.RGBA{R: 200, G: 50, B: 0, A: 255}    // Example enemy color
	enemyBodyColor     = color.RGBA{R: 255, G: 80, B: 0, A: 255}    // Example enemy color
	foodStandardColor  = color.RGBA{R: 255, G: 0, B: 0, A: 255}     // Red
	foodSpeedColor     = color.RGBA{R: 255, G: 165, B: 0, A: 255}   // Orange
	foodSlowColor      = color.RGBA{R: 0, G: 191, B: 255, A: 255}   // Deep Sky Blue
	foodFlashColor     = color.RGBA{R: 255, G: 255, B: 200, A: 180} // Pale yellow flash
	speedUpColorShift  = color.RGBA{R: 255, G: 100, B: 100, A: 80}  // Reddish tint overlay
	slowDownColorShift = color.RGBA{R: 100, G: 100, B: 255, A: 80}  // Bluish tint overlay
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

	// 5. Draw Effects (e.g., food flash) - Draw before snakes
	drawEffects(screen, state)

	// 6. Draw Enemy Snakes
	for _, enemy := range state.EnemySnakes {
		if enemy != nil {
			// TODO: Pass effect state if enemies have speed effects
			drawSnake(screen, *enemy, assets)
		}
	}

	// 7. Draw Player Snake (drawn last to be on top)
	if state.PlayerSnake != nil {
		drawSnake(screen, *state.PlayerSnake, assets)
	}

	// 7. Draw HUD (Score, etc.) - To be implemented later
	drawHUD(screen, state.Score)
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

// drawSnake draws a single snake using sprites with interpolation and effects.
func drawSnake(screen *ebiten.Image, s game.Snake, assets *assets.Manager) {
	if len(s.Body) == 0 || len(s.PrevBody) == 0 || len(s.Body) != len(s.PrevBody) || assets.SnakeBody == nil || assets.SnakeHead == nil {
		// log.Printf("DrawSnake skip: BodyLen=%d, PrevBodyLen=%d, BodyAsset=%v, HeadAsset=%v", len(s.Body), len(s.PrevBody), assets.SnakeBody, assets.SnakeHead)
		return // Cannot draw without assets or consistent body/prevBody
	}

	bodyW, bodyH := assets.SnakeBody.Size()
	headW, headH := assets.SnakeHead.Size()
	progress := s.MoveProgress // How far we are into the current move (0.0 to < 1.0)

	// Helper function for linear interpolation
	lerp := func(a, b float64, t float64) float64 {
		return a + (b-a)*t
	}

	// Check for active speed effect
	var speedEffectColor color.Color = nil
	if !s.SpeedEffectEndTime.IsZero() && time.Now().Before(s.SpeedEffectEndTime) {
		if s.SpeedFactor > 1.0 {
			speedEffectColor = speedUpColorShift
		} else if s.SpeedFactor < 1.0 {
			speedEffectColor = slowDownColorShift
		}
	}

	// Draw segments (Body and Head)
	for i := 0; i < len(s.Body); i++ {
		segment := s.Body[i]
		prevSegmentPos := s.PrevBody[i]
		visX := lerp(float64(prevSegmentPos.X), float64(segment.X), progress)
		visY := lerp(float64(prevSegmentPos.Y), float64(segment.Y), progress)

		var img *ebiten.Image
		var imgW, imgH int
		var angle float64 = 0
		op := &ebiten.DrawImageOptions{}

		if i == 0 { // Head
			img = assets.SnakeHead
			imgW, imgH = headW, headH // Already got size earlier
			// Calculate head rotation based on logical direction
			switch s.Direction {
			case game.DirUp:
				angle = -math.Pi / 2
			case game.DirDown:
				angle = math.Pi / 2
			case game.DirLeft:
				angle = math.Pi
			case game.DirRight:
				angle = 0
			}
		} else { // Body
			img = assets.SnakeBody
			imgW, imgH = bodyW, bodyH // Already got size earlier
			// Calculate body rotation based on visual segment connection
			segmentInFront := s.Body[i-1]
			prevSegmentInFront := s.PrevBody[i-1]
			visFrontX := lerp(float64(prevSegmentInFront.X), float64(segmentInFront.X), progress)
			visFrontY := lerp(float64(prevSegmentInFront.Y), float64(segmentInFront.Y), progress)
			dx := visFrontX - visX
			dy := visFrontY - visY
			if math.Abs(dx) < 0.01 {
				angle = math.Pi / 2
			} else if math.Abs(dy) < 0.01 {
				angle = 0
			} else {
				angle = math.Atan2(dy, dx) /* Optional: Snap? */
			}
		}

		// Common Drawing Logic
		tx := visX*float64(GridCellSize) + float64(GridCellSize-imgW)/2.0
		ty := visY*float64(GridCellSize) + float64(GridCellSize-imgH)/2.0
		centerX := float64(imgW) / 2.0
		centerY := float64(imgH) / 2.0
		op.GeoM.Translate(-centerX, -centerY)
		op.GeoM.Rotate(angle)
		op.GeoM.Translate(centerX, centerY)
		op.GeoM.Translate(tx, ty)

		// Apply speed effect color modification if active
		if speedEffectColor != nil {
			op.ColorScale.ScaleWithColor(speedEffectColor) // Use ColorScale for tinting
		}

		screen.DrawImage(img, op)
	}
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

// drawEffects renders transient visual effects.
func drawEffects(screen *ebiten.Image, state game.RenderableState) {
	// Food Eaten Flash - REMOVED
	/*
		if state.FoodEatenPos != nil {
			// Simple square flash effect
			fx := float32(state.FoodEatenPos.X * GridCellSize)
			fy := float32(state.FoodEatenPos.Y * GridCellSize)
			size := float32(GridCellSize) // Flash covers the cell
			vector.DrawFilledRect(screen, fx, fy, size, size, foodFlashColor, false)
		}
	*/

	// TODO: Add spawning effects
	// TODO: Add collision effects
}

// drawHUD function renders the Heads-Up Display (Score, etc.)
func drawHUD(screen *ebiten.Image, score int /*, other hud data */) {
	scoreStr := fmt.Sprintf("Score: %d", score)

	// Simple text rendering at top-left. Improve with fonts later.
	// Use ebitenutil which we should have imported.
	ebitenutil.DebugPrintAt(screen, scoreStr, 10, 10)

	// TODO: Add rendering for speed effect duration if needed
}
