package assets

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Asset paths (relative to the executable or run command location)
const (
	imgDir = "internal/assets/images"
)

// Manager handles loading and storing assets.
type Manager struct {
	// Images
	SnakeHead    *ebiten.Image
	SnakeBody    *ebiten.Image
	FoodStandard *ebiten.Image
	FoodSpeedUp  *ebiten.Image
	FoodSlowDown *ebiten.Image
	Background   *ebiten.Image
	Wall         *ebiten.Image

	// Add maps for sounds later
}

// NewManager creates and loads assets.
func NewManager() (*Manager, error) {
	m := &Manager{}
	var err error

	// Load Images
	m.SnakeHead, err = loadImage("head.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load head image: %w", err)
	}
	m.SnakeBody, err = loadImage("body.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load body image: %w", err)
	}
	m.FoodStandard, err = loadImage("food1.png") // Example mapping
	if err != nil {
		return nil, fmt.Errorf("failed to load food1 image: %w", err)
	}
	m.FoodSpeedUp, err = loadImage("food2.png") // Example mapping
	if err != nil {
		return nil, fmt.Errorf("failed to load food2 image: %w", err)
	}
	m.FoodSlowDown, err = loadImage("food3.png") // Example mapping
	if err != nil {
		return nil, fmt.Errorf("failed to load food3 image: %w", err)
	}

	// Load optional assets (handle potential errors gracefully)
	m.Background, err = loadImage("background.png")
	if err != nil {
		log.Printf("Warning: Failed to load background image: %v", err)
		m.Background = nil // Allow game to run without it
	}
	m.Wall, err = loadImage("wall.png")
	if err != nil {
		log.Printf("Warning: Failed to load wall image: %v", err)
		m.Wall = nil // Use default drawing if wall sprite fails
	}

	log.Println("Assets loaded successfully.")
	return m, nil
}

// loadImage is a helper to load an image from the assets directory.
func loadImage(name string) (*ebiten.Image, error) {
	path := filepath.Join(imgDir, name)
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", path, err)
	}
	return img, nil
}
