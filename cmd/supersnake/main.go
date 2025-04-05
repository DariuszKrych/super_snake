package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"snake-game/internal/game" // Reference game constants
	"snake-game/internal/scene"
	"snake-game/internal/scene/gameover" // Import gameover scene
	"snake-game/internal/scene/gameplay" // Import gameplay scene

	// Import other scenes (MainMenu, Pause, etc.) when created
	"snake-game/internal/render" // Import render package
)

const (
	// Keep screen dimensions consistent for now
	// We can make this more dynamic later if needed
	screenWidth  = game.GridWidth * render.GridCellSize  // Use render constant
	screenHeight = game.GridHeight * render.GridCellSize // Use render constant
)

func main() {
	// Seed random number generator once at the start
	rand.Seed(time.Now().UnixNano())

	// Create the scene manager
	manager := scene.NewManager(screenWidth, screenHeight)

	// --- Register Scenes ---
	// Register Gameplay Scene
	manager.RegisterScene(scene.SceneTypeGameplay, func() scene.Scene { return gameplay.NewGameplayScene() })
	// Register MainMenu Scene (when created)
	// manager.RegisterScene(scene.SceneTypeMainMenu, func() scene.Scene { return mainmenu.NewMainMenuScene() })
	// Register GameOver Scene
	manager.RegisterScene(scene.SceneTypeGameOver, func() scene.Scene { return gameover.NewGameOverScene() })
	// Register Pause Scene (when created)
	// manager.RegisterScene(scene.SceneTypePause, func() scene.Scene { return pause.NewPauseScene() })

	// --- Set Initial Scene ---
	manager.SetInitialScene(scene.SceneTypeGameplay) // Start with Gameplay for now
	// manager.SetInitialScene(scene.SceneTypeMainMenu) // Change this to MainMenu later

	// Configure Ebitengine window
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Super Snake GO")
	// ebiten.SetFullscreen(true) // Disable fullscreen for now during development
	ebiten.SetFullscreen(true) // Re-enable fullscreen

	// Run the game using the SceneManager as the ebiten.Game implementation
	if err := ebiten.RunGame(manager); err != nil {
		log.Fatalf("Ebitengine RunGame error: %v", err)
	}
}
