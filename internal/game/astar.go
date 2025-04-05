// internal/game/astar.go
package game

import (
	"container/heap"
)

// --- A* Pathfinding Implementation ---

// aStarNode represents a node in the A* search space.
type aStarNode struct {
	pos    Position   // Grid position
	g      int        // Cost from start to node
	h      int        // Heuristic cost from node to target
	f      int        // g + h
	parent *aStarNode // Parent node for path reconstruction
	index  int        // Index in the priority queue
}

// --- Priority Queue Implementation (Min-Heap based on f-cost) ---

type priorityQueue []*aStarNode

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// Min-heap based on f-cost, tie-break with h-cost (optional)
	if pq[i].f == pq[j].f {
		return pq[i].h < pq[j].h
	}
	return pq[i].f < pq[j].f
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*aStarNode)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an item in the queue.
func (pq *priorityQueue) update(item *aStarNode, g, h int) {
	item.g = g
	item.h = h
	item.f = g + h
	heap.Fix(pq, item.index)
}

// --- A* Helper Functions ---

// heuristic calculates the Manhattan distance.
func heuristic(a, b Position) int {
	dx := a.X - b.X
	dy := a.Y - b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// isValid checks if a position is within grid boundaries.
func isValid(pos Position, width, height int) bool {
	return pos.X >= 0 && pos.X < width && pos.Y >= 0 && pos.Y < height
}

// reconstructPath builds the path from the target node back to the start.
// Returns a slice of Positions, excluding the start, including the target.
func reconstructPath(targetNode *aStarNode) []Position {
	path := []Position{}
	current := targetNode
	for current != nil && current.parent != nil { // Stop before adding the start node's position
		path = append(path, current.pos)
		current = current.parent
	}
	// Reverse the path
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// findPath implements the A* algorithm.
func findPath(start, target Position, width, height int, obstacles map[Position]bool) []Position {
	openSet := make(priorityQueue, 0)
	heap.Init(&openSet)

	closedSet := make(map[Position]bool)
	nodeMap := make(map[Position]*aStarNode) // To quickly find existing nodes

	startNode := &aStarNode{pos: start, g: 0, h: heuristic(start, target)}
	startNode.f = startNode.g + startNode.h
	heap.Push(&openSet, startNode)
	nodeMap[start] = startNode

	// Define neighbors relative positions (no diagonals)
	neighbors := []Position{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}}

	for openSet.Len() > 0 {
		current := heap.Pop(&openSet).(*aStarNode)

		if current.pos == target {
			return reconstructPath(current)
		}

		closedSet[current.pos] = true

		for _, offset := range neighbors {
			neighborPos := Position{X: current.pos.X + offset.X, Y: current.pos.Y + offset.Y}

			// Check bounds, obstacles, and if already processed
			if !isValid(neighborPos, width, height) || obstacles[neighborPos] || closedSet[neighborPos] {
				continue
			}

			tentativeG := current.g + 1 // Cost of moving to neighbor is 1

			neighborNode, exists := nodeMap[neighborPos]
			if !exists {
				neighborNode = &aStarNode{
					pos:    neighborPos,
					parent: current,
				}
				nodeMap[neighborPos] = neighborNode
				heap.Push(&openSet, neighborNode)
				// Set costs directly here as it's the first time seeing the node
				neighborNode.g = tentativeG
				neighborNode.h = heuristic(neighborPos, target)
				neighborNode.f = neighborNode.g + neighborNode.h
				heap.Fix(&openSet, neighborNode.index) // Need to fix after setting costs
			} else if tentativeG < neighborNode.g {
				// Found a better path to this existing node
				neighborNode.parent = current
				openSet.update(neighborNode, tentativeG, heuristic(neighborPos, target))
			}
		}
	}

	return nil // No path found
}
