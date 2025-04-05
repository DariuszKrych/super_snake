package game

import (
	"math/rand"
	"time"
)

// --- Constants ---

const (
	GridWidth         = 40
	GridHeight        = 30
	InitialSpeed      = 8
	SpeedIncrement    = 0.5
	MaxSpeed          = 20
	InitialSnakeLen   = 3
	InitialFoodItems  = 3               // Start with this many food items
	MaxTotalFoodItems = 50              // Maximum food items on screen
	FoodSpawnInterval = 5 * time.Second // Time between new food spawns
)

// --- Types ---

// Direction represents movement direction
type Direction int

const (
	DirNone Direction = iota
	DirUp
	DirDown
	DirLeft
	DirRight
)

// Position represents a point on the grid
type Position struct {
	X, Y int
}

// Snake struct holds state for a single snake (player or AI)
type Snake struct {
	Body        []Position
	Direction   Direction
	NextDir     Direction   // Buffer for next direction input
	SpeedFactor float64     // Multiplier for speed (1.0 = normal, >1 = faster, <1 = slower)
	SpeedTimer  *time.Timer // Timer for temporary speed effects
	IsPlayer    bool        // Flag to distinguish player snake
	// Add other snake-specific properties if needed (e.g., color for rendering)
}

// FoodType defines the kind of food
type FoodType int

const (
	FoodTypeStandard FoodType = iota
	FoodTypeSpeedUp
	FoodTypeSlowDown
)

// Food struct holds state for a food item
type Food struct {
	Pos      Position
	Type     FoodType
	Points   int
	Effect   func(*Snake)  // Function to apply the food's effect
	Duration time.Duration // Duration for temporary effects
	// Add rendering-specific info later (e.g., sprite name)
}

// Game struct holds the entire game state
type Game struct {
	PlayerSnake       *Snake
	EnemySnakes       []*Snake
	FoodItems         []*Food
	Score             int
	Speed             float64
	MoveTicker        *time.Ticker
	IsOver            bool
	IsPaused          bool
	nextFoodSpawnTime time.Time // When the next food item should appear
}

// --- Game Initialization ---

// NewGame initializes a new game state
func NewGame() *Game {
	g := &Game{
		Speed:     InitialSpeed,
		FoodItems: make([]*Food, 0, 5), // Initialize with some capacity
	}
	g.Reset()
	return g
}

// Reset initializes or resets the game state for a new round
func (g *Game) Reset() {
	// Initialize player snake
	startX, startY := GridWidth/2, GridHeight/2
	initialBody := make([]Position, InitialSnakeLen)
	for i := 0; i < InitialSnakeLen; i++ {
		initialBody[i] = Position{X: startX - i, Y: startY}
	}
	g.PlayerSnake = &Snake{
		Body:        initialBody,
		Direction:   DirRight,
		NextDir:     DirRight,
		SpeedFactor: 1.0,
		IsPlayer:    true,
	}

	// TODO: Initialize Enemy Snakes (Section 5.4)
	g.EnemySnakes = []*Snake{} // Empty for now

	g.Score = 0
	g.Speed = InitialSpeed
	g.IsOver = false
	g.IsPaused = false
	g.FoodItems = g.FoodItems[:0] // Clear existing food

	// Spawn initial food items
	for i := 0; i < InitialFoodItems; i++ {
		g.spawnFoodItem()
	}

	// Schedule the next timed spawn
	g.scheduleNextFoodSpawn()

	if g.MoveTicker != nil {
		g.MoveTicker.Stop()
	}
	if g.PlayerSnake != nil && g.PlayerSnake.SpeedFactor != 0 {
		g.MoveTicker = time.NewTicker(time.Second / time.Duration(g.Speed*g.PlayerSnake.SpeedFactor))
	} else {
		g.MoveTicker = time.NewTicker(time.Second / time.Duration(g.Speed))
	}

	// TODO: Reset enemy states
}

// --- Food Logic ---

func (g *Game) scheduleNextFoodSpawn() {
	// Add some randomness to the interval if desired
	// interval := FoodSpawnInterval + time.Duration(rand.Intn(2000)) * time.Millisecond
	interval := FoodSpawnInterval
	g.nextFoodSpawnTime = time.Now().Add(interval)
}

// spawnFoodItem places a *single* food item randomly, avoiding obstacles.
func (g *Game) spawnFoodItem() {
	// Check if maximum food limit is reached
	if len(g.FoodItems) >= MaxTotalFoodItems {
		return
	}

	occupied := make(map[Position]bool)
	if g.PlayerSnake != nil {
		for _, segment := range g.PlayerSnake.Body {
			occupied[segment] = true
		}
	}
	for _, enemy := range g.EnemySnakes {
		if enemy != nil {
			for _, segment := range enemy.Body {
				occupied[segment] = true
			}
		}
	}
	for _, food := range g.FoodItems {
		if food != nil {
			occupied[food.Pos] = true
		}
	}

	// Determine food type based on probability (Section 5.5)
	foodType := FoodTypeStandard // Default
	points := 10
	var effect func(*Snake) = nil
	duration := 0 * time.Second
	r := rand.Float64()
	if r < 0.15 {
		foodType = FoodTypeSpeedUp
	} else if r < 0.30 {
		foodType = FoodTypeSlowDown
	}
	switch foodType {
	case FoodTypeStandard:
		points = 10
		effect = func(s *Snake) { s.grow() }
	case FoodTypeSpeedUp:
		points = 15
		duration = 7 * time.Second
		effect = func(s *Snake) { s.grow(); s.applySpeedBoost(1.5, duration) }
	case FoodTypeSlowDown:
		points = 5
		duration = 7 * time.Second
		effect = func(s *Snake) { s.grow(); s.applySpeedBoost(0.6, duration) }
	}

	// Find an empty spot
	var newPos Position
	attempts := 0
	maxAttempts := GridWidth*GridHeight - len(occupied)
	if maxAttempts <= 0 {
		return
	} // No space left

	for attempts < maxAttempts*2 { // Allow more attempts for sparse grids
		newPos = Position{X: rand.Intn(GridWidth), Y: rand.Intn(GridHeight)}
		if !occupied[newPos] {
			break
		}
		attempts++
	}

	if occupied[newPos] {
		return
	} // Could not find a spot

	newItem := &Food{
		Pos:      newPos,
		Type:     foodType,
		Points:   points,
		Effect:   effect,
		Duration: duration,
	}
	g.FoodItems = append(g.FoodItems, newItem)
}

// --- Snake Logic ---

// grow increases snake length by duplicating the tail segment
func (s *Snake) grow() {
	if len(s.Body) == 0 {
		return // Should not happen
	}
	tail := s.Body[len(s.Body)-1]
	s.Body = append(s.Body, tail) // Duplicate tail segment
}

// applySpeedBoost applies a temporary speed multiplier
func (s *Snake) applySpeedBoost(factor float64, duration time.Duration) {
	// Stop existing timer if any
	if s.SpeedTimer != nil {
		s.SpeedTimer.Stop()
	}

	s.SpeedFactor = factor
	// TODO: Need to notify the main game loop to adjust the MoveTicker interval

	s.SpeedTimer = time.AfterFunc(duration, func() {
		s.SpeedFactor = 1.0
		s.SpeedTimer = nil
		// TODO: Notify main game loop to reset MoveTicker interval
	})
}

// move calculates the next head position and updates the body
// Returns true if food was eaten
func (s *Snake) move() bool {
	if len(s.Body) == 0 {
		return false
	}

	// Update direction based on buffered input
	s.Direction = s.NextDir

	// Calculate new head position
	head := s.Body[0]
	newHead := head
	switch s.Direction {
	case DirUp:
		newHead.Y--
	case DirDown:
		newHead.Y++
	case DirLeft:
		newHead.X--
	case DirRight:
		newHead.X++
	}

	// Update snake body: Prepend new head, remove tail (unless grown)
	// Growth is handled separately when food is confirmed eaten
	s.Body = append([]Position{newHead}, s.Body[:len(s.Body)-1]...)

	return false // Food eating check happens in the main game update
}

// checkCollision checks if the snake's head collides with boundaries or itself
func (s *Snake) checkCollision(width, height int) (hitWall bool, hitSelf bool) {
	if len(s.Body) == 0 {
		return false, false
	}
	head := s.Body[0]

	// Check boundary collision
	if head.X < 0 || head.X >= width || head.Y < 0 || head.Y >= height {
		return true, false
	}

	// Check self collision (check against body segments from index 1 onwards)
	for i := 1; i < len(s.Body); i++ {
		if head == s.Body[i] {
			return false, true
		}
	}

	return false, false
}

// --- Game Update Logic ---

// Update proceeds the game state by one tick
func (g *Game) Update() error {
	if g.IsOver || g.IsPaused {
		return nil
	}

	// Check if food needs spawning based on time interval
	if time.Now().After(g.nextFoodSpawnTime) {
		g.spawnFoodItem()
		g.scheduleNextFoodSpawn() // Schedule the next one
	}

	// Check the ticker to see if it's time to move player
	select {
	case <-g.MoveTicker.C:
		g.updatePlayer()
		// TODO: Update Enemy AI snakes
	default:
		// Not time to move yet
	}

	// TODO: Update Enemy AI movements & checks
	g.updateEnemies()

	return nil
}

// updatePlayer handles player snake movement and collision checks
func (g *Game) updatePlayer() {
	if g.PlayerSnake == nil || len(g.PlayerSnake.Body) == 0 {
		return
	}

	headBeforeMove := g.PlayerSnake.Body[0]
	ateFoodIndex := -1 // Index of the food item eaten

	// Check food collision *before* moving the body segments
	for i, food := range g.FoodItems {
		if food != nil && headBeforeMove == food.Pos {
			ateFoodIndex = i
			g.Score += food.Points
			if food.Effect != nil {
				food.Effect(g.PlayerSnake)
			}
			break // Eat only one food per tick
		}
	}

	// Remove eaten food item (if any)
	if ateFoodIndex != -1 {
		g.FoodItems = append(g.FoodItems[:ateFoodIndex], g.FoodItems[ateFoodIndex+1:]...)
		// Immediately try to spawn a replacement food item
		g.spawnFoodItem()
	}

	// Move player snake body
	g.PlayerSnake.move()

	// Check player collisions after moving
	hitWall, hitSelf := g.PlayerSnake.checkCollision(GridWidth, GridHeight)
	if hitWall || hitSelf {
		g.triggerGameOver("Player Collision")
		return
	}

	// Check collision with enemy snakes (Section 5.4)
	if g.checkPlayerEnemyCollisions() {
		return
	}
}

// updateEnemies handles AI snake movement and collision checks
func (g *Game) updateEnemies() {
	// TODO: Implement AI update loop (Section 5.4)
	// - Get AI input/decision (pathfinding, avoidance)
	// - Move AI snakes
	// - Check AI collisions (wall, self, other AI, player body)
	// - Handle AI eating food
	// - Remove dead AI snakes
}

// checkPlayerEnemyCollisions handles interactions between player and enemies
func (g *Game) checkPlayerEnemyCollisions() bool {
	// TODO: Implement player vs enemy collision logic (Section 5.4)
	// - Player head vs Enemy body -> Player dies
	// - Player head vs Enemy head -> Both die (Player game over)
	// - Enemy head vs Player body -> Enemy dies
	return false // Placeholder
}

// triggerGameOver sets the game over state
func (g *Game) triggerGameOver(reason string) {
	// TODO: Add reason handling if needed
	g.IsOver = true
	if g.MoveTicker != nil {
		g.MoveTicker.Stop()
	}
	// Stop any active speed timers
	if g.PlayerSnake != nil && g.PlayerSnake.SpeedTimer != nil {
		g.PlayerSnake.SpeedTimer.Stop()
	}
	// TODO: Play Game Over sound
}

// TogglePause pauses or resumes the game
func (g *Game) TogglePause() {
	g.IsPaused = !g.IsPaused
	// TODO: Adjust ticker/timers when pausing/resuming
	if g.IsPaused {
		if g.MoveTicker != nil {
			g.MoveTicker.Stop() // Stop movement ticker
		}
		// TODO: Pause SpeedTimers if implemented precisely
	} else {
		// Resume: Reset ticker based on current speed
		if g.MoveTicker != nil {
			g.MoveTicker.Reset(time.Second / time.Duration(g.Speed*g.PlayerSnake.SpeedFactor))
		}
		// TODO: Resume SpeedTimers
	}
}

// HandleInput updates the player's next direction based on input
func (g *Game) HandleInput(newDir Direction) {
	// Prevent immediate reversal
	currentDir := g.PlayerSnake.Direction
	isValidMove := true
	switch newDir {
	case DirUp:
		if currentDir == DirDown {
			isValidMove = false
		}
	case DirDown:
		if currentDir == DirUp {
			isValidMove = false
		}
	case DirLeft:
		if currentDir == DirRight {
			isValidMove = false
		}
	case DirRight:
		if currentDir == DirLeft {
			isValidMove = false
		}
	}

	if isValidMove {
		g.PlayerSnake.NextDir = newDir
	}
}

// GetState provides necessary info for rendering
type RenderableState struct {
	PlayerSnake         *Snake
	EnemySnakes         []*Snake
	FoodItems           []*Food
	Score               int
	IsOver              bool
	IsPaused            bool
	GridWidth           int
	GridHeight          int
	PlayerSpeedFactor   float64
	SpeedEffectDuration time.Duration
}

func (g *Game) GetState() RenderableState {
	var remainingDuration time.Duration

	playerSnakeCopy := g.PlayerSnake
	// Create a copy of the food slice to avoid modification during rendering
	foodItemsCopy := make([]*Food, len(g.FoodItems))
	copy(foodItemsCopy, g.FoodItems)

	speedFactor := 1.0
	if playerSnakeCopy != nil {
		speedFactor = playerSnakeCopy.SpeedFactor
	}

	return RenderableState{
		PlayerSnake:         playerSnakeCopy,
		EnemySnakes:         g.EnemySnakes,
		FoodItems:           foodItemsCopy, // Return the slice
		Score:               g.Score,
		IsOver:              g.IsOver,
		IsPaused:            g.IsPaused,
		GridWidth:           GridWidth,
		GridHeight:          GridHeight,
		PlayerSpeedFactor:   speedFactor,
		SpeedEffectDuration: remainingDuration,
	}
}
