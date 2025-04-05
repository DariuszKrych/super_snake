package gameplay

import (
	"log"

	"snake-game/internal/game"
	"snake-game/internal/input"
	"snake-game/internal/render"
	"snake-game/internal/scene"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// GameplayScene holds the state for the main gameplay.
type GameplayScene struct {
	gameData *game.Game
	inputMgr *input.Manager
	sceneMgr scene.ManagerInterface
	// Add specific rendering assets or state if needed
}

// NewGameplayScene creates a new gameplay scene instance.
func NewGameplayScene() *GameplayScene {
	return &GameplayScene{}
}

// Load initializes the scene.
func (s *GameplayScene) Load(manager scene.ManagerInterface, gameData *game.Game) {
	log.Println("Loading Gameplay Scene")
	s.sceneMgr = manager
	s.inputMgr = manager.GetInputManager() // Get shared input manager
	s.gameData = gameData
	// Reset game state if necessary when loading gameplay
	s.gameData.Reset()
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
		s.gameData.HandleInput(dir) // Update player's intended direction
	}

	switch action {
	case input.ActionPause:
		s.gameData.TogglePause() // Toggle pause state in game logic
		// Optionally, transition to a PauseScene
		// return scene.Transition{FromScene: scene.SceneTypeGameplay, ToScene: scene.SceneTypePause}, nil
	case input.ActionConfirm:
		// Maybe used for something else in gameplay?
	case input.ActionRestart:
		// Maybe add a dedicated restart key
		s.gameData.Reset()
	}

	// 2. Update Game Logic (if not paused)
	if !s.gameData.IsPaused {
		err := s.gameData.Update() // This updates snake movement, checks collisions, spawns food
		if err != nil {
			return scene.Transition{}, err // Propagate errors
		}
	}

	// 3. Check for Game Over state change
	if s.gameData.IsOver {
		// Transition to GameOver scene
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

	// Draw Pause overlay if paused
	if s.gameData.IsPaused {
		// TODO: Implement a nicer pause overlay
		width, height := s.sceneMgr.GetWindowSize()
		ebitenutil.DebugPrintAt(screen, "PAUSED (Press P/Esc to Resume)", width/2-100, height/2)
	}
}
