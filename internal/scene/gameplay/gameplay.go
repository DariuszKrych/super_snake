package gameplay

import (
	"image/color"
	"log"

	"snake-game/internal/game"
	"snake-game/internal/input"
	"snake-game/internal/particle"
	"snake-game/internal/render"
	"snake-game/internal/scene"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// GameplayScene holds the state for the main gameplay.
type GameplayScene struct {
	gameData    *game.Game
	inputMgr    *input.Manager
	sceneMgr    scene.ManagerInterface
	particleSys *particle.System
	// Add specific rendering assets or state if needed
}

// NewGameplayScene creates a new gameplay scene instance.
func NewGameplayScene() *GameplayScene {
	ps := particle.NewSystem(0)
	return &GameplayScene{
		particleSys: ps,
	}
}

// Load initializes the scene.
func (s *GameplayScene) Load(manager scene.ManagerInterface, gameData *game.Game) {
	log.Println("Loading Gameplay Scene")
	s.sceneMgr = manager
	s.inputMgr = manager.GetInputManager()
	s.gameData = gameData
	s.gameData.Reset()
	s.particleSys.Particles = s.particleSys.Particles[:0]
	// Load gameplay-specific assets here (e.g., sounds)
}

// Unload cleans up the scene.
func (s *GameplayScene) Unload() scene.SceneType {
	log.Println("Unloading Gameplay Scene")
	// Unload assets
	return scene.SceneTypeGameplay
}

// Update handles game logic updates.
func (s *GameplayScene) Update(manager scene.ManagerInterface) (scene.Transition, error) {
	// 1. Handle Input
	dir, action := s.inputMgr.Update()

	if dir != game.DirNone {
		s.gameData.HandleInput(dir)
	}

	switch action {
	case input.ActionPause:
		s.gameData.TogglePause()
	case input.ActionConfirm:
	case input.ActionRestart:
		s.gameData.Reset()
		s.particleSys.Particles = s.particleSys.Particles[:0]
	}

	// Update particle system
	deltaTime := 1.0 / float64(ebiten.TPS())
	s.particleSys.Update(deltaTime)

	// 2. Update Game Logic (if not paused)
	if !s.gameData.IsPaused {
		err := s.gameData.Update(deltaTime)
		if err != nil {
			return scene.Transition{}, err
		}

		// Check if food was eaten by PLAYER
		lastPlayerEatenPos := s.gameData.FoodEatenPos
		if lastPlayerEatenPos != nil {
			flashColor := color.RGBA{R: 255, G: 255, B: 180, A: 255}
			centerX := float64(lastPlayerEatenPos.X*render.GridCellSize) + float64(render.GridCellSize)/2.0
			centerY := float64(lastPlayerEatenPos.Y*render.GridCellSize) + float64(render.GridCellSize)/2.0
			s.particleSys.Emit(particle.EmitConfig{
				X:              centerX,
				Y:              centerY,
				Count:          15,
				UseGravity:     false,
				Color:          flashColor,
				VelocitySpread: 80,
				MinLifetime:    0.2,
				MaxLifetime:    0.5,
				MinSize:        1,
				MaxSize:        3,
			})
			// s.gameData.FoodEatenPos = nil // Game logic now clears this based on time
		}

		// Check if food was eaten by ENEMY
		lastEnemyEatenPos := s.gameData.EnemyFoodEatenPos
		if lastEnemyEatenPos != nil {
			flashColor := color.RGBA{R: 255, G: 180, B: 180, A: 255} // Different color for enemy eat
			centerX := float64(lastEnemyEatenPos.X*render.GridCellSize) + float64(render.GridCellSize)/2.0
			centerY := float64(lastEnemyEatenPos.Y*render.GridCellSize) + float64(render.GridCellSize)/2.0
			s.particleSys.Emit(particle.EmitConfig{
				X:              centerX,
				Y:              centerY,
				Count:          10, // Fewer particles for enemy?
				UseGravity:     false,
				Color:          flashColor,
				VelocitySpread: 60,
				MinLifetime:    0.15,
				MaxLifetime:    0.4,
				MinSize:        1,
				MaxSize:        2,
			})
			s.gameData.EnemyFoodEatenPos = nil // Consume the event signal here
		}
	}

	// 3. Check for Game Over state change
	if s.gameData.IsOver {
		return scene.Transition{FromScene: scene.SceneTypeGameplay, ToScene: scene.SceneTypeGameOver}, nil
	}

	// No transition requested
	return scene.Transition{}, nil
}

// Draw renders the gameplay screen.
func (s *GameplayScene) Draw(screen *ebiten.Image) {
	// Get the current renderable state from the game logic
	renderState := s.gameData.GetState()
	// Get assets from the scene manager
	assets := s.sceneMgr.GetAssets()

	// Use the render package to draw everything, passing assets
	render.DrawGame(screen, renderState, assets)

	// Draw particles on top
	s.particleSys.Draw(screen)

	// Draw Pause overlay if paused
	if s.gameData.IsPaused {
		width, height := s.sceneMgr.GetWindowSize()
		ebitenutil.DebugPrintAt(screen, "PAUSED (Press P/Esc to Resume)", width/2-100, height/2)
	}
}
