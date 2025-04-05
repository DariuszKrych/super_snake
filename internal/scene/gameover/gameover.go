package gameover

import (
	"fmt"
	"image/color"
	"log"

	"snake-game/internal/game"
	"snake-game/internal/input"
	"snake-game/internal/scene"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// GameOverScene displays the game over message and score.
type GameOverScene struct {
	sceneMgr   scene.ManagerInterface
	inputMgr   *input.Manager
	finalScore int
	// Add assets like fonts if needed
}

// NewGameOverScene creates a new game over scene instance.
// We might pass the final score here eventually.
func NewGameOverScene() *GameOverScene {
	return &GameOverScene{}
}

// Load initializes the scene.
func (s *GameOverScene) Load(manager scene.ManagerInterface, gameData *game.Game) {
	log.Println("Loading GameOver Scene")
	s.sceneMgr = manager
	s.inputMgr = manager.GetInputManager()
	s.finalScore = gameData.Score // Get score from the ended game state
	// Load assets if needed
}

// Unload cleans up the scene.
func (s *GameOverScene) Unload() scene.SceneType {
	log.Println("Unloading GameOver Scene")
	// Unload assets
	return scene.SceneTypeGameOver
}

// Update handles input for restarting or exiting.
func (s *GameOverScene) Update(manager scene.ManagerInterface) (scene.Transition, error) {
	_, action := s.inputMgr.Update()

	switch action {
	case input.ActionConfirm: // Typically Space or Enter
		// Transition back to Gameplay (which will call Reset)
		return scene.Transition{FromScene: scene.SceneTypeGameOver, ToScene: scene.SceneTypeGameplay}, nil
	case input.ActionBack: // Typically Escape
		// TODO: Implement transition to Main Menu or Exit
		log.Println("Exit/Back action from GameOver not implemented yet.")
	}

	// No transition requested
	return scene.Transition{}, nil
}

// Draw renders the game over screen.
func (s *GameOverScene) Draw(screen *ebiten.Image) {
	width, height := s.sceneMgr.GetWindowSize()

	// Simple background overlay (optional)
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	ebitenutil.DrawRect(screen, 0, 0, float64(width), float64(height), overlayColor)

	// Game Over Text
	title := "GAME OVER"
	scoreMsg := fmt.Sprintf("Final Score: %d", s.finalScore)
	prompt := "Press Space/Enter to Restart"

	// Basic text rendering (Improve with actual fonts later)
	titleX := (width - len(title)*8) / 2
	scoreX := (width - len(scoreMsg)*8) / 2
	promptX := (width - len(prompt)*8) / 2

	titleY := height/2 - 30
	scoreY := height / 2
	promptY := height/2 + 30

	ebitenutil.DebugPrintAt(screen, title, titleX, titleY)
	ebitenutil.DebugPrintAt(screen, scoreMsg, scoreX, scoreY)
	ebitenutil.DebugPrintAt(screen, prompt, promptX, promptY)
}
