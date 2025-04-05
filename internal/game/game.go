package game

import (
	// Need heap for astar.go (if not already imported)
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
	currentPath        []Position  // Path for AI snakes
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
		currentPath:        nil,
	}

	// Initialize Enemies
	g.EnemySnakes = make([]*Snake, 0, MaxEnemySnakes)
	for i := 0; i < NumEnemySnakes; i++ {
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
				currentPath:        nil,
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

// updateEnemyAI uses A* pathfinding to set NextDir.
func (g *Game) updateEnemyAI(s *Snake) {
	if len(s.Body) == 0 {
		return
	}
	head := s.Body[0]

	// --- Path Following ---
	if len(s.currentPath) > 0 {
		// Check if the next step in the path is the current head position
		// This can happen if the path calculation was slightly delayed
		if s.currentPath[0] == head {
			s.currentPath = s.currentPath[1:] // Pop the current head position
			if len(s.currentPath) == 0 {
				// Reached end of path, need to recalculate
				goto recalculate // Use goto for clarity in this state machine
			}
		}

		// Set NextDir based on the first step in the existing path
		nextStep := s.currentPath[0]
		newDir := directionFromTo(head, nextStep)
		if newDir != DirNone {
			// Basic check: don't immediately reverse into self
			canMove := true
			if len(s.Body) > 1 {
				neck := s.Body[1]
				potentialNextHead := head
				switch newDir {
				case DirUp:
					potentialNextHead.Y--
				case DirDown:
					potentialNextHead.Y++
				case DirLeft:
					potentialNextHead.X--
				case DirRight:
					potentialNextHead.X++
				}
				if potentialNextHead == neck {
					canMove = false
					// log.Printf("AI %p avoiding neck collision by recalculating", s)
					s.currentPath = nil // Invalidate path, force recalculation
					goto recalculate
				}
			}
			if canMove {
				s.NextDir = newDir
				return // Successfully following path
			}
		}
	}

recalculate: // Label for jumping to path recalculation
	// --- Path Recalculation ---
	targetFood := g.findClosestFood(head)
	if targetFood == nil {
		g.setRandomEnemyDirection(s) // No food, move randomly
		return
	}

	// Build obstacle map
	obstacles := g.buildObstacleMap(s) // Exclude self head

	// Find path
	path := findPath(head, targetFood.Pos, GridWidth, GridHeight, obstacles)

	if path != nil && len(path) > 0 {
		s.currentPath = path
		// Set direction based on the first step
		newDir := directionFromTo(head, path[0])
		if newDir != DirNone {
			s.NextDir = newDir
		} else {
			// Should not happen if path is valid
			log.Printf("Warning: A* path resulted in invalid first step for AI %p", s)
			g.setRandomEnemyDirection(s) // Fallback
		}
	} else {
		// No path found (food unreachable or blocked)
		// log.Printf("AI %p could not find path to food at %v", s, targetFood.Pos)
		g.setRandomEnemyDirection(s) // Fallback: Move randomly but avoid obstacles
	}
}

// findClosestFood finds the nearest food item to a given position.
func (g *Game) findClosestFood(pos Position) *Food {
	var closestFood *Food = nil
	minDist := -1

	for _, food := range g.FoodItems {
		if food == nil {
			continue
		}
		dist := heuristic(pos, food.Pos) // Manhattan distance
		if closestFood == nil || dist < minDist {
			minDist = dist
			closestFood = food
		}
	}
	return closestFood
}

// buildObstacleMap creates a map of all occupied cells for pathfinding.
// Includes wall padding and all snake segments except the head of the snake `self`.
func (g *Game) buildObstacleMap(self *Snake) map[Position]bool {
	obstacles := make(map[Position]bool)

	// Player Snake Body (Include head now for avoidance)
	if g.PlayerSnake != nil {
		// for i, seg := range g.PlayerSnake.Body {
		// 	if i > 0 { // Skip player head
		// 		obstacles[seg] = true
		// 	}
		// }
		for _, seg := range g.PlayerSnake.Body {
			obstacles[seg] = true // Include player head as obstacle
		}
	}

	// Other Enemy Snakes (include head and body)
	for _, enemy := range g.EnemySnakes {
		if enemy != nil && enemy != self {
			for _, seg := range enemy.Body {
				obstacles[seg] = true
			}
		}
	}

	// Self Body (excluding head)
	if self != nil {
		for i, seg := range self.Body {
			if i > 0 { // Still exclude self head
				obstacles[seg] = true
			}
		}
	}

	// TODO: Add walls as obstacles explicitly if needed for A*?
	// Currently relies on isValid check, might be slightly less efficient.

	return obstacles
}

// setRandomEnemyDirection chooses a valid random direction, avoiding immediate obstacles.
func (g *Game) setRandomEnemyDirection(s *Snake) {
	head := s.Body[0]
	possibleDirs := []Direction{DirUp, DirDown, DirLeft, DirRight}
	validDirs := []Direction{}

	obstacles := g.buildObstacleMap(s) // Need current obstacles

	for _, dir := range possibleDirs {
		// Prevent immediate reversal
		if (dir == DirUp && s.Direction == DirDown) || (dir == DirDown && s.Direction == DirUp) ||
			(dir == DirLeft && s.Direction == DirRight) || (dir == DirRight && s.Direction == DirLeft) {
			continue
		}

		// Check if the next cell is valid and not an obstacle
		nextPos := head
		switch dir {
		case DirUp:
			nextPos.Y--
		case DirDown:
			nextPos.Y++
		case DirLeft:
			nextPos.X--
		case DirRight:
			nextPos.X++
		}
		if isValid(nextPos, GridWidth, GridHeight) && !obstacles[nextPos] {
			validDirs = append(validDirs, dir)
		}
	}

	if len(validDirs) > 0 {
		s.NextDir = validDirs[rand.Intn(len(validDirs))]
	} else {
		// Nowhere to go? Keep current direction (will likely collide)
		s.NextDir = s.Direction
		// log.Printf("AI %p trapped! No valid random move.", s)
	}
	s.currentPath = nil // Clear path as we are moving randomly
}

// directionFromTo calculates the direction needed to move from pos 'from' to pos 'to'.
func directionFromTo(from, to Position) Direction {
	if to.Y < from.Y {
		return DirUp
	}
	if to.Y > from.Y {
		return DirDown
	}
	if to.X < from.X {
		return DirLeft
	}
	if to.X > from.X {
		return DirRight
	}
	return DirNone // Should not happen for adjacent cells
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
		s.MoveProgress -= 1.0
		// Clear the path step we just took *if* we were following one
		if !s.IsPlayer && len(s.currentPath) > 0 && s.currentPath[0] == s.Body[0] { // Compare with head *before* updating Body
			// Pop the step we just completed
			s.currentPath = s.currentPath[1:]
		}

		// 1. Finalize the move for this step
		// Store current body as previous body (needs a deep copy)
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
