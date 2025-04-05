package game

import (
	"log"
	"math/rand"
	"time"
	// Import log for debugging if needed
	// "log"
)

// --- Constants ---

const (
	GridWidth          = 40
	GridHeight         = 30
	InitialSpeed       = 8 // Grid cells per second
	SpeedIncrement     = 0.5
	MaxSpeed           = 20
	InitialSnakeLen    = 3
	InitialFoodItems   = 3                // Start with this many food items
	MaxTotalFoodItems  = 50               // Maximum food items on screen
	FoodSpawnInterval  = 5 * time.Second  // Time between new food spawns
	NumEnemySnakes     = 2                // Initial number of enemies
	MaxEnemySnakes     = 3                // Maximum number of enemies allowed
	EnemySpawnInterval = 15 * time.Second // Time between trying to spawn new enemies
	foodFlashDuration  = 150 * time.Millisecond
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
	PlayerSnake        *Snake
	EnemySnakes        []*Snake
	FoodItems          []*Food
	Score              int
	Speed              float64 // Base grid cells per second for player
	IsOver             bool
	IsPaused           bool
	nextFoodSpawnTime  time.Time // When the next food item should appear
	nextEnemySpawnTime time.Time // When to next check for enemy spawning
	FoodEatenPos       *Position // Position where food was last eaten
	FoodEatenTime      time.Time // Time when food was last eaten
	EnemyFoodEatenPos  *Position // Position where an enemy last ate food
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
	occupied := make(map[Position]bool) // Track occupied spots during init

	// Initialize player snake
	startX, startY := GridWidth/4, GridHeight/2 // Start player on left side
	initialBody := make([]Position, InitialSnakeLen)
	prevBody := make([]Position, InitialSnakeLen)
	for i := 0; i < InitialSnakeLen; i++ {
		pos := Position{X: startX - i, Y: startY}
		initialBody[i] = pos
		prevBody[i] = pos
		occupied[pos] = true
	}
	g.PlayerSnake = &Snake{
		Body:               initialBody,
		PrevBody:           prevBody,
		Direction:          DirRight,
		NextDir:            DirRight,
		SpeedFactor:        1.0,
		SpeedEffectEndTime: time.Time{},
		IsPlayer:           true,
		MoveProgress:       0.0,
	}

	// Initialize Enemies
	g.EnemySnakes = make([]*Snake, 0, MaxEnemySnakes) // Use Max for capacity
	for i := 0; i < NumEnemySnakes; i++ {             // Start with initial number
		enemy := g.createEnemy(occupied)
		if enemy != nil {
			g.EnemySnakes = append(g.EnemySnakes, enemy)
			for _, seg := range enemy.Body {
				occupied[seg] = true
			}
		}
	}

	g.Score = 0
	g.Speed = InitialSpeed
	g.IsOver = false
	g.IsPaused = false
	g.FoodItems = g.FoodItems[:0] // Clear existing food
	g.FoodEatenPos = nil          // Reset food eaten effect tracker
	g.FoodEatenTime = time.Time{}
	g.EnemyFoodEatenPos = nil // Reset enemy food effect tracker

	// Spawn initial food items (avoiding snakes)
	for i := 0; i < InitialFoodItems; i++ {
		g.spawnFoodItem()
	}

	g.scheduleNextFoodSpawn()
	g.scheduleNextEnemySpawn() // Schedule first enemy spawn check
}

// createEnemy initializes a single enemy snake at a valid position.
func (g *Game) createEnemy(occupied map[Position]bool) *Snake {
	attempts := 0
	maxAttempts := (GridWidth * GridHeight) / 2 // Limit attempts

	for attempts < maxAttempts {
		// Try placing on the right side initially
		startX := GridWidth - GridWidth/4 + rand.Intn(GridWidth/4)
		startY := rand.Intn(GridHeight)
		startDir := DirLeft // Start moving left

		// Check if start position + initial body is clear
		validPlacement := true
		tempBody := make([]Position, InitialSnakeLen)
		for i := 0; i < InitialSnakeLen; i++ {
			// Calculate initial body based on startDir (simplified: assumes left)
			pos := Position{X: startX + i, Y: startY}
			if occupied[pos] || pos.X >= GridWidth || pos.X < 0 || pos.Y >= GridHeight || pos.Y < 0 {
				validPlacement = false
				break
			}
			tempBody[i] = pos
		}

		if validPlacement {
			initialBody := make([]Position, InitialSnakeLen)
			prevBody := make([]Position, InitialSnakeLen)
			for i := 0; i < InitialSnakeLen; i++ {
				pos := Position{X: startX + i, Y: startY}
				initialBody[i] = pos
				prevBody[i] = pos
			}
			return &Snake{
				Body:               initialBody,
				PrevBody:           prevBody,
				Direction:          startDir,
				NextDir:            startDir,
				SpeedFactor:        1.0, // Enemies move at base speed for now
				SpeedEffectEndTime: time.Time{},
				IsPlayer:           false,
				MoveProgress:       0.0,
			}
		}
		attempts++
	}
	log.Printf("Warning: Could not place enemy snake after %d attempts", maxAttempts)
	return nil // Failed to place enemy
}

// --- Food Logic ---

func (g *Game) scheduleNextFoodSpawn() {
	// Add some randomness to the interval if desired
	// interval := FoodSpawnInterval + time.Duration(rand.Intn(2000)) * time.Millisecond
	interval := FoodSpawnInterval
	g.nextFoodSpawnTime = time.Now().Add(interval)
}

// scheduleNextEnemySpawn sets the time for the next enemy spawn check.
func (g *Game) scheduleNextEnemySpawn() {
	g.nextEnemySpawnTime = time.Now().Add(EnemySpawnInterval)
}

// spawnFoodItem places a *single* food item randomly, avoiding obstacles.
func (g *Game) spawnFoodItem() {
	if len(g.FoodItems) >= MaxTotalFoodItems {
		return
	}
	occupied := make(map[Position]bool)
	// Populate occupied map (include player AND enemies)
	if g.PlayerSnake != nil {
		for _, seg := range g.PlayerSnake.Body {
			occupied[seg] = true
		}
	}
	for _, enemy := range g.EnemySnakes {
		if enemy != nil {
			for _, seg := range enemy.Body {
				occupied[seg] = true
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
func (g *Game) Update(deltaTime float64) error {
	if g.IsOver || g.IsPaused {
		return nil
	}

	// Check timed food spawning
	if time.Now().After(g.nextFoodSpawnTime) {
		g.spawnFoodItem()
		g.scheduleNextFoodSpawn()
	}

	// Check timed enemy spawning
	if time.Now().After(g.nextEnemySpawnTime) {
		g.spawnEnemyIfPossible()
		g.scheduleNextEnemySpawn() // Schedule next check regardless of success
	}

	// Update Player Snake Movement Progress
	if g.PlayerSnake != nil {
		g.updateSnakeProgress(g.PlayerSnake, deltaTime)
		if g.IsOver {
			return nil // Stop updates if player died this frame
		}
	}

	// Update Enemy AI Movement Progress
	// Iterate backwards for safe removal
	for i := len(g.EnemySnakes) - 1; i >= 0; i-- {
		enemy := g.EnemySnakes[i]
		if enemy != nil {
			g.updateEnemyAI(enemy) // Determine NextDir for enemy
			g.updateSnakeProgress(enemy, deltaTime)
			if g.IsOver {
				return nil // Stop if player died colliding with this enemy
			}
		}
	}

	return nil
}

// updateEnemyAI sets the NextDir for an enemy snake (placeholder logic).
func (g *Game) updateEnemyAI(s *Snake) {
	// Very basic AI: Try to move towards the first food item.
	// Does not avoid obstacles yet!
	if len(g.FoodItems) > 0 && g.FoodItems[0] != nil {
		target := g.FoodItems[0].Pos
		head := s.Body[0]
		currentDir := s.Direction

		// Determine desired direction (simplified)
		var desiredDir Direction
		if target.X > head.X && currentDir != DirLeft {
			desiredDir = DirRight
		} else if target.X < head.X && currentDir != DirRight {
			desiredDir = DirLeft
		} else if target.Y > head.Y && currentDir != DirUp {
			desiredDir = DirDown
		} else if target.Y < head.Y && currentDir != DirDown {
			desiredDir = DirUp
		} else {
			// If already aligned on one axis, move on the other, or keep current
			if target.Y != head.Y && currentDir != DirUp && currentDir != DirDown {
				if target.Y > head.Y {
					desiredDir = DirDown
				} else {
					desiredDir = DirUp
				}
			} else if target.X != head.X && currentDir != DirLeft && currentDir != DirRight {
				if target.X > head.X {
					desiredDir = DirRight
				} else {
					desiredDir = DirLeft
				}
			} else {
				desiredDir = currentDir // Keep going if blocked or unsure
			}
		}

		// Basic anti-reversal check
		if (desiredDir == DirUp && currentDir == DirDown) ||
			(desiredDir == DirDown && currentDir == DirUp) ||
			(desiredDir == DirLeft && currentDir == DirRight) ||
			(desiredDir == DirRight && currentDir == DirLeft) {
			// If desired move is direct reverse, try to turn perpendicular instead (simple avoidance)
			if currentDir == DirUp || currentDir == DirDown {
				desiredDir = DirLeft
			} else {
				desiredDir = DirUp
			}
			// TODO: Need better obstacle check here
		}

		s.NextDir = desiredDir
	} else {
		// No food? Move randomly or keep going (for now, keep going)
		s.NextDir = s.Direction
	}
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
				if s.IsPlayer {
					g.Score += food.Points
				}
				if food.Effect != nil {
					food.Effect(s) // Apply effect (which might call s.grow())
				}
				// Immediately try to spawn replacement
				g.spawnFoodItem()

				// Trigger food eaten effect
				pos := food.Pos // Copy position
				if s.IsPlayer {
					g.FoodEatenPos = &pos
					g.FoodEatenTime = time.Now()
				} else {
					g.EnemyFoodEatenPos = &pos // Set enemy signal
				}

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
				g.triggerGameOver("Player Self/Wall Collision")
			} else {
				g.removeEnemySnake(s) // Remove enemy on collision
			}
			return // Stop processing this snake if it died
		}

		// Check Inter-snake Collisions
		if g.checkInterSnakeCollisions(s) {
			// If this snake died or caused player game over, stop processing it.
			if g.IsOver || !g.isSnakeAlive(s) {
				return
			}
		}
	}
}

// isSnakeAlive checks if a given snake pointer still exists in the EnemySnakes slice.
// Used after collision checks to see if the snake was removed.
func (g *Game) isSnakeAlive(snake *Snake) bool {
	if snake.IsPlayer {
		return true // Player handled by g.IsOver
	}
	for _, enemy := range g.EnemySnakes {
		if enemy == snake {
			return true
		}
	}
	return false
}

// checkInterSnakeCollisions checks collisions between the given snake `s` and all other snakes.
// Returns true if a collision occurred that requires stopping processing for `s`.
func (g *Game) checkInterSnakeCollisions(s *Snake) bool {
	if len(s.Body) == 0 {
		return false
	}
	head := s.Body[0]

	// Check against player if `s` is an enemy
	if !s.IsPlayer && g.PlayerSnake != nil && len(g.PlayerSnake.Body) > 0 {
		playerHead := g.PlayerSnake.Body[0]
		// Head-on check
		if head == playerHead {
			g.triggerGameOver("Enemy Head-on Collision")
			g.removeEnemySnake(s)
			return true // Player game over, stop processing enemy
		}
		// Check if enemy head hit player body
		for i := 1; i < len(g.PlayerSnake.Body); i++ {
			if head == g.PlayerSnake.Body[i] {
				g.removeEnemySnake(s)
				// TODO: Award points?
				return true // Enemy died, stop processing it
			}
		}
	}

	// Check against enemies
	for _, other := range g.EnemySnakes {
		if s == other || other == nil || len(other.Body) == 0 {
			continue // Skip self and dead enemies
		}
		otherHead := other.Body[0]

		// Head-on check (Enemy vs Enemy or Player vs Enemy)
		if head == otherHead {
			if s.IsPlayer {
				g.triggerGameOver("Player Head-on Collision")
				g.removeEnemySnake(other)
				return true // Player game over
			} else {
				// Both enemies die
				g.removeEnemySnake(s)
				g.removeEnemySnake(other)
				return true // Current enemy `s` died
			}
		}

		// Check if `s` head hit `other` body
		for i := 1; i < len(other.Body); i++ {
			if head == other.Body[i] {
				if s.IsPlayer {
					g.triggerGameOver("Player Hit Enemy Body")
					return true // Player game over
				} else {
					// Enemy hit another enemy's body
					g.removeEnemySnake(s)
					return true // Current enemy `s` died
				}
			}
		}
	}
	return false // No relevant collision found for `s`
}

// removeEnemySnake removes a specific enemy snake from the game slice.
func (g *Game) removeEnemySnake(snakeToRemove *Snake) {
	newEnemyList := g.EnemySnakes[:0]
	for _, s := range g.EnemySnakes {
		if s != snakeToRemove {
			newEnemyList = append(newEnemyList, s)
		} else {
			log.Printf("Enemy snake removed due to collision.")
			// TODO: Trigger enemy death effect/sound
		}
	}
	g.EnemySnakes = newEnemyList
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
	EnemyFoodEatenPos   *Position
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

	// Clear player food eaten effect if duration passed
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
		EnemyFoodEatenPos:   g.EnemyFoodEatenPos,
	}
}

// spawnEnemyIfPossible attempts to add a new enemy if below the max count.
func (g *Game) spawnEnemyIfPossible() {
	if len(g.EnemySnakes) < MaxEnemySnakes {
		log.Printf("Attempting to spawn new enemy snake (current: %d)", len(g.EnemySnakes))
		// Need to gather all currently occupied positions
		occupied := make(map[Position]bool)
		if g.PlayerSnake != nil {
			for _, seg := range g.PlayerSnake.Body {
				occupied[seg] = true
			}
		}
		for _, enemy := range g.EnemySnakes {
			if enemy != nil {
				for _, seg := range enemy.Body {
					occupied[seg] = true
				}
			}
		}
		for _, food := range g.FoodItems {
			if food != nil {
				occupied[food.Pos] = true
			}
		}

		newEnemy := g.createEnemy(occupied)
		if newEnemy != nil {
			g.EnemySnakes = append(g.EnemySnakes, newEnemy)
			log.Printf("New enemy snake spawned (total: %d)", len(g.EnemySnakes))
		} else {
			log.Printf("Failed to spawn new enemy snake (could not find placement).")
		}
	}
}
