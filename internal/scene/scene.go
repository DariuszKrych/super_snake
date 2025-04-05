package scene

import (
	"snake-game/internal/game"  // Import our game logic package
	"snake-game/internal/input" // Import input package

	"github.com/hajimehoshi/ebiten/v2"
)

// Transition represents a request to change to a different scene.
type Transition struct {
	FromScene SceneType
	ToScene   SceneType
	// Add any data needed for the transition (e.g., final score for GameOver)
}

// SceneType identifies different scenes in the game.
type SceneType int

const (
	SceneTypeUndefined SceneType = iota
	SceneTypeMainMenu
	SceneTypeGameplay
	SceneTypeGameOver
	SceneTypePause
	// Add SceneTypeOptions if needed
)

// ManagerInterface defines the methods a scene manager needs.
// Scenes will use this to request transitions.
type ManagerInterface interface {
	GoTo(transition Transition)
	GetWindowSize() (int, int)
	GetInputManager() *input.Manager
	// Add methods for accessing shared resources like assets if needed
}

// Scene defines the interface that all game scenes must implement.
type Scene interface {
	// Update handles logic updates for the scene.
	// It returns a Transition request if the scene should change, otherwise nil.
	// It also returns an error if something goes wrong.
	Update(manager ManagerInterface) (Transition, error)

	// Draw renders the scene to the screen.
	Draw(screen *ebiten.Image)

	// Load is called when the scene becomes active (optional initialization).
	Load(manager ManagerInterface, gameData *game.Game)

	// Unload is called when the scene is replaced (optional cleanup).
	Unload() SceneType // Returns the type of the scene being unloaded
}

// SceneConstructor is a function type that creates a new scene.
type SceneConstructor func() Scene
