package scene

import (
	"fmt"
	"log"

	"snake-game/internal/assets" // Import assets package
	"snake-game/internal/game"   // Import our core game logic
	"snake-game/internal/input"  // Import the input package

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	// "snake-game/internal/scene/gameplay" // Remove this import
	// "snake-game/internal/scene/mainmenu"
)

// Manager handles scene transitions and holds the current scene.
type Manager struct {
	current           Scene
	nextScene         Scene // Scene to transition to
	transition        *Transition
	screenWidth       int
	screenHeight      int
	gameData          *game.Game                     // Shared game state data
	inputManager      *input.Manager                 // Add input manager instance
	assetManager      *assets.Manager                // Add asset manager instance
	sceneConstructors map[SceneType]SceneConstructor // Map to store scene constructors
	// Add asset managers, input managers etc. here if needed globally
}

// NewManager creates a new scene manager and loads assets.
func NewManager(screenWidth, screenHeight int) *Manager {
	// Load assets first
	assetMgr, err := assets.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize asset manager: %v", err)
	}

	m := &Manager{
		screenWidth:       screenWidth,
		screenHeight:      screenHeight,
		gameData:          game.NewGame(),     // Initialize the core game data
		inputManager:      input.NewManager(), // Initialize the input manager
		assetManager:      assetMgr,           // Store the loaded assets
		sceneConstructors: make(map[SceneType]SceneConstructor),
	}
	// Scenes must be registered before being used.
	// Registration will happen in main or an init function.

	// Start with a default scene type (e.g., MainMenu or Gameplay)
	// Ensure the scene is registered before calling GoTo or setting m.current directly
	// m.GoTo(Transition{ToScene: SceneTypeGameplay}) // Example transition
	// For now, set a placeholder until registration is done
	m.current = NewPlaceholderScene(SceneTypeUndefined) // Start with undefined placeholder

	return m
}

// RegisterScene adds a scene constructor to the manager.
func (m *Manager) RegisterScene(sceneType SceneType, constructor SceneConstructor) {
	if _, exists := m.sceneConstructors[sceneType]; exists {
		log.Printf("Warning: Scene type %v already registered. Overwriting.", sceneType)
	}
	m.sceneConstructors[sceneType] = constructor
	log.Printf("Registered scene type %v", sceneType)
}

// SetInitialScene sets the first scene to be loaded.
// Should be called after registering scenes.
func (m *Manager) SetInitialScene(sceneType SceneType) {
	constructor, exists := m.sceneConstructors[sceneType]
	if !exists {
		log.Fatalf("Error: Initial scene type %v not registered!", sceneType)
	}
	m.current = constructor()
	m.current.Load(m, m.gameData)
	log.Printf("Set initial scene to %v", sceneType)
}

// Update updates the current scene and handles transitions.
func (m *Manager) Update() error {
	if m.transition != nil {
		// Unload old scene
		if m.current != nil {
			m.current.Unload()
		}
		// Set and load new scene
		m.current = m.nextScene
		if m.current != nil {
			m.current.Load(m, m.gameData)
		}
		// Reset transition state
		m.nextScene = nil
		m.transition = nil
	}

	if m.current != nil {
		transitionReq, err := m.current.Update(m)
		if err != nil {
			return fmt.Errorf("error updating scene %T: %w", m.current, err)
		}
		if (transitionReq != Transition{}) { // Check if a valid transition was requested
			m.GoTo(transitionReq)
		}
	}
	return nil
}

// Draw draws the current scene.
func (m *Manager) Draw(screen *ebiten.Image) {
	if m.current != nil {
		m.current.Draw(screen)
	}
}

// Layout is required by ebiten.Game interface.
func (m *Manager) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Store the actual window size if needed, but return configured logical size
	// m.screenWidth = outsideWidth
	// m.screenHeight = outsideHeight
	return m.screenWidth, m.screenHeight
}

// GoTo initiates a scene transition.
func (m *Manager) GoTo(transition Transition) {
	if m.transition != nil {
		log.Printf("Warning: Already transitioning from %v to %v, ignoring request to go to %v", m.transition.FromScene, m.transition.ToScene, transition.ToScene)
		return
	}

	constructor, exists := m.sceneConstructors[transition.ToScene]
	if !exists {
		log.Printf("Error: Scene type %v not registered for transition", transition.ToScene)
		return // Cancel transition if scene doesn't exist
	}

	log.Printf("Transition requested from %v to %v", transition.FromScene, transition.ToScene)
	m.transition = &transition
	m.nextScene = constructor() // Use the constructor to create the scene instance

	// Removed the old switch statement that directly instantiated scenes
}

// GetWindowSize returns the logical screen dimensions.
func (m *Manager) GetWindowSize() (int, int) {
	return m.screenWidth, m.screenHeight
}

// GetInputManager returns the shared input manager.
// Scenes can call this via the ManagerInterface.
func (m *Manager) GetInputManager() *input.Manager {
	return m.inputManager
}

// GetAssets returns the shared asset manager.
func (m *Manager) GetAssets() *assets.Manager {
	return m.assetManager
}

// --- Placeholder Scene --- (Keep for GameOver/Pause for now)

type PlaceholderScene struct {
	sceneType SceneType
}

func NewPlaceholderScene(t SceneType) *PlaceholderScene {
	return &PlaceholderScene{sceneType: t}
}

func (s *PlaceholderScene) Update(manager ManagerInterface) (Transition, error) {
	// No update logic for placeholder
	return Transition{}, nil
}

func (s *PlaceholderScene) Draw(screen *ebiten.Image) {
	// Simple placeholder drawing
	msg := fmt.Sprintf("Placeholder Scene: %v", s.sceneType)
	ebitenutil.DebugPrintAt(screen, msg, 10, 10)
}

func (s *PlaceholderScene) Load(manager ManagerInterface, gameData *game.Game) {
	log.Printf("Loading Placeholder Scene: %v", s.sceneType)
}

func (s *PlaceholderScene) Unload() SceneType {
	log.Printf("Unloading Placeholder Scene: %v", s.sceneType)
	return s.sceneType
}
