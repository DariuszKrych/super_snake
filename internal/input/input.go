package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"snake-game/internal/game" // For game.Direction
)

// Action represents a game action triggered by input.
type Action int

const (
	ActionNone Action = iota
	ActionMoveUp
	ActionMoveDown
	ActionMoveLeft
	ActionMoveRight
	ActionPause
	ActionConfirm // e.g., for menus
	ActionBack    // e.g., for menus
	ActionRestart
)

// Manager handles reading input state.
type Manager struct {
	// We could add configuration here later, e.g., key bindings
}

// NewManager creates a new input manager.
func NewManager() *Manager {
	return &Manager{}
}

// Update checks the current input state and returns relevant actions/directions.
// This simple version directly returns the first detected movement direction.
// A more complex game might queue actions.
func (m *Manager) Update() (game.Direction, Action) {
	// Check for movement keys (prioritize arrows, then WASD)
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		return game.DirUp, ActionNone
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		return game.DirDown, ActionNone
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		return game.DirLeft, ActionNone
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		return game.DirRight, ActionNone
	}

	// Check for action keys
	if inpututil.IsKeyJustPressed(ebiten.KeyP) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		// Use Escape primarily for pausing during gameplay, maybe backing out of menus
		return game.DirNone, ActionPause // For now, map both to Pause
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		// Use Space primarily for restarting when game over, Enter for menu confirm
		return game.DirNone, ActionConfirm // For now, map both to Confirm
	}
	// Add ActionRestart check if needed (e.g., R key)
	// Add ActionBack check if needed (e.g., Backspace or specific key for menus)

	return game.DirNone, ActionNone // No relevant input detected
}
