package game

import (
	"math/rand"
	"time"
	// Import log for debugging if needed
	// "log"
)

// --- Constants ---

const (
	GridWidth         = 40
	GridHeight        = 30
	InitialSpeed      = 8 // Grid cells per second
	SpeedIncrement    = 0.5
	MaxSpeed          = 20
	InitialSnakeLen   = 3
	InitialFoodItems  = 3               // Start with this many food items
	MaxTotalFoodItems = 50              // Maximum food items on screen
	FoodSpawnInterval = 5 * time.Second // Time between new food spawns
	foodFlashDuration = 150 * time.Millisecond
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
	Body               []Position
	PrevBody           []Position // Stores body positions from the *previous completed* move step
	Direction          Direction
	NextDir            Direction   // Buffer for next direction input
	SpeedFactor        float64     // Multiplier for speed (1.0 = normal, >1 = faster, <1 = slower)
	SpeedTimer         *time.Timer // Timer for temporary speed effects
	SpeedEffectEndTime time.Time   // Track when the speed boost ends
	IsPlayer           bool        // Flag to distinguish player snake
	MoveProgress       float64     // How far into the current grid move (0.0 to 1.0)
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
	Speed             float64 // Base grid cells per second for player
	IsOver            bool
	IsPaused          bool
	nextFoodSpawnTime time.Time // When the next food item should appear
	FoodEatenPos      *Position // Position where food was last eaten
	FoodEatenTime     time.Time // Time when food was last eaten
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
	startX, startY := GridWidth/2, GridHeight/2
	initialBody := make([]Position, InitialSnakeLen)
	prevBody := make([]Position, InitialSnakeLen)
	for i := 0; i < InitialSnakeLen; i++ {
		pos := Position{X: startX - i, Y: startY}
		initialBody[i] = pos
		prevBody[i] = pos // Initially, previous position is the same
	}
	g.PlayerSnake = &Snake{
		Body:               initialBody,
		PrevBody:           prevBody,
		Direction:          DirRight,
		NextDir:            DirRight,
		SpeedFactor:        1.0,
		SpeedEffectEndTime: time.Time{}, // Ensure effect time is zeroed
		IsPlayer:           true,
		MoveProgress:       0.0, // Start not moving
	}

	// TODO: Initialize Enemy Snakes (Section 5.4)
	g.EnemySnakes = []*Snake{} // Empty for now

	g.Score = 0
	g.Speed = InitialSpeed
	g.IsOver = false
	g.IsPaused = false
	g.FoodItems = g.FoodItems[:0] // Clear existing food
	g.FoodEatenPos = nil          // Reset food eaten effect tracker
	g.FoodEatenTime = time.Time{}

	// Spawn initial food items
	for i := 0; i < InitialFoodItems; i++ {
		g.spawnFoodItem()
	}

	// Schedule the next timed spawn
	g.scheduleNextFoodSpawn()
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
// Needs to update both Body and PrevBody
func (s *Snake) grow() {
	if len(s.Body) == 0 {
		return
	}
	tail := s.Body[len(s.Body)-1]
	s.Body = append(s.Body, tail)
	// Also append to PrevBody using the *current* last segment of PrevBody
	if len(s.PrevBody) > 0 {
		prevTail := s.PrevBody[len(s.PrevBody)-1]
		s.PrevBody = append(s.PrevBody, prevTail)
	} else {
		// Handle edge case if PrevBody is empty but Body isn't (shouldn't happen)
		s.PrevBody = append(s.PrevBody, tail) // Use Body's tail as fallback
	}
}

// applySpeedBoost applies a temporary speed multiplier
func (s *Snake) applySpeedBoost(factor float64, duration time.Duration) {
	if s.SpeedTimer != nil {
		s.SpeedTimer.Stop()
	}
	s.SpeedFactor = factor
	endTime := time.Now().Add(duration) // Calculate end time
	s.SpeedEffectEndTime = endTime      // Store end time
	s.SpeedTimer = time.AfterFunc(duration, func() {
		s.SpeedFactor = 1.0
		s.SpeedTimer = nil
		s.SpeedEffectEndTime = time.Time{} // Reset end time
	})
}

// checkCollision checks if the snake's head collides with boundaries or itself
// This is checked *only* when a move is finalized.
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

// Update proceeds the game state by one frame
func (g *Game) Update(deltaTime float64) error { // Accept delta time
	if g.IsOver || g.IsPaused {
		return nil
	}

	// Check timed food spawning
	if time.Now().After(g.nextFoodSpawnTime) {
		g.spawnFoodItem()
		g.scheduleNextFoodSpawn()
	}

	// Update Player Snake Movement Progress
	if g.PlayerSnake != nil {
		g.updateSnakeProgress(g.PlayerSnake, deltaTime)
	}

	// TODO: Update Enemy AI Movement Progress
	// for _, enemy := range g.EnemySnakes {
	// 	g.updateSnakeProgress(enemy, deltaTime)
	// }

	return nil
}

// updateSnakeProgress handles movement progress and finalization for a single snake
func (g *Game) updateSnakeProgress(s *Snake, deltaTime float64) {
	if len(s.Body) == 0 {
		return
	}

	// Calculate movement amount for this frame
	moveAmount := s.SpeedFactor * g.Speed * deltaTime
	s.MoveProgress += moveAmount

	// Did the snake complete one or more grid moves this frame?
	for s.MoveProgress >= 1.0 {
		s.MoveProgress -= 1.0 // Consume one full grid move

		// 1. Finalize the move for this step
		// Store current body as previous body (needs a deep copy)
		// s.PrevBody = s.Body // Incorrect: this assigns the slice header, not data
		s.PrevBody = make([]Position, len(s.Body))
		copy(s.PrevBody, s.Body)

		// Determine actual direction for this step
		s.Direction = s.NextDir

		// Calculate next head position
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

		// Check for food at the *target* position *before* updating body
		ateFoodIndex := -1
		for i, food := range g.FoodItems {
			if food != nil && newHead == food.Pos {
				ateFoodIndex = i
				g.Score += food.Points
				if food.Effect != nil {
					food.Effect(s) // Apply effect (which might call s.grow())
				}
				// Immediately try to spawn replacement
				g.spawnFoodItem()

				// Trigger food eaten effect
				pos := food.Pos // Copy position
				g.FoodEatenPos = &pos
				g.FoodEatenTime = time.Now()

				break
			}
		}

		// Update body: Prepend new head, potentially grow
		if ateFoodIndex != -1 {
			// Body grew inside food.Effect(), just prepend new head
			// Need to ensure grow() updated both Body and PrevBody correctly
			newBody := make([]Position, len(s.Body))
			newBody[0] = newHead
			copy(newBody[1:], s.Body[:len(s.Body)-1]) // Shift old body
			s.Body = newBody

			// Remove eaten food *after* potential growth
			g.FoodItems = append(g.FoodItems[:ateFoodIndex], g.FoodItems[ateFoodIndex+1:]...)
		} else {
			// No food eaten, normal move: Prepend new head, drop tail
			newBody := make([]Position, len(s.Body))
			newBody[0] = newHead
			copy(newBody[1:], s.Body[:len(s.Body)-1])
			s.Body = newBody
		}

		// 2. Check Collisions (only after finalizing position)
		hitWall, hitSelf := s.checkCollision(GridWidth, GridHeight)
		if hitWall || hitSelf {
			if s.IsPlayer {
				g.triggerGameOver("Player Collision")
			} else {
				// TODO: Handle enemy death
			}
			return // Stop processing this snake if it died
		}

		// TODO: Check inter-snake collisions (Player vs Enemy, Enemy vs Enemy)
		if s.IsPlayer && g.checkPlayerEnemyCollisions() {
			return // Game Over handled in checkPlayerEnemyCollisions
		}
		// Add enemy collision checks here too
	}
}

// updatePlayer is now removed, logic is in updateSnakeProgress
/* func (g *Game) updatePlayer() { ... } */

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
		if g.PlayerSnake != nil && g.PlayerSnake.SpeedTimer != nil {
			g.PlayerSnake.SpeedTimer.Stop()
		}
		// TODO: Pause SpeedTimers if implemented precisely
	} else {
		// Resume: Reset ticker based on current speed
		if g.PlayerSnake != nil && g.PlayerSnake.SpeedTimer != nil {
			g.PlayerSnake.SpeedTimer.Reset(time.Second / time.Duration(g.Speed*g.PlayerSnake.SpeedFactor))
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

// GetState provides necessary info for rendering, including progress
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
	FoodEatenPos        *Position
	FoodEatenTime       time.Time
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

	// Clear food eaten effect if duration passed
	if g.FoodEatenPos != nil && time.Since(g.FoodEatenTime) > foodFlashDuration {
		g.FoodEatenPos = nil
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
		FoodEatenPos:        g.FoodEatenPos,
		FoodEatenTime:       g.FoodEatenTime,
	}
}
