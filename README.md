# Super Snake GO

An enhanced Snake game built with Golang and the Ebitengine 2D library.

## Current Status

This is a work-in-progress implementation based on detailed requirements. Current features include:

*   **Core Gameplay:** Player-controlled snake (Arrows/WASD), movement, growth on eating food, game over on wall/self collision.
*   **Food System:**
    *   Multiple food items appear on screen (starts with 3, max 50).
    *   New food items spawn every 5 seconds.
    *   Eating a food item immediately spawns a replacement.
    *   Different food types implemented (Standard, Speed-Up, Slow-Down) with point differences and temporary speed effects.
*   **Scene Management:** Basic structure with transitions between Gameplay and Game Over scenes.
*   **Rendering:** Simple rectangle-based graphics for snake, food, and walls using Ebitengine.
*   **Input:** Handles snake movement, pause (P/Esc), and restart from Game Over (Space/Enter).
*   **Platform:** Runs fullscreen on Linux (and potentially macOS/Windows with Ebitengine's cross-platform support).

## Requirements

*   **Go:** Version 1.18 or later recommended.
*   **Ebitengine Dependencies (Ubuntu/Debian):**
    ```bash
    sudo apt-get update && sudo apt-get install -y build-essential libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev libxxf86vm-dev libasound2-dev libglfw3 libglfw3-dev
    ```
    *(See Ebitengine documentation for dependencies on other platforms.)*

## How to Build

Navigate to the project's root directory (`super_snake`) in your terminal and run:

```bash
go build ./cmd/supersnake/
```

This will create an executable named `supersnake` (or `supersnake.exe` on Windows) in the project root.

## How to Run

1.  **Build:** Follow the build instructions above.
2.  **Run:** Execute the created binary from the project root:
    ```bash
    ./supersnake
    ```

Alternatively, run directly without building a separate executable:

```bash
go run ./cmd/supersnake/main.go
```

## Controls

*   **Move:** Arrow Keys or WASD keys
*   **Pause/Resume:** `P` or `Escape`
*   **Restart (Game Over Screen):** `Space` or `Enter`

## Project Structure

*   `cmd/supersnake/`: Main application entry point.
*   `internal/`: Contains core packages:
    *   `game/`: Core game logic (snake, food, state, rules).
    *   `scene/`: Scene interface, manager, and specific scenes (`gameplay/`, `gameover/`).
    *   `input/`: Input handling.
    *   `render/`: Rendering logic.
    *   (Placeholder directories: `assets/`, `audio/`, `ai/`)

## Next Steps / TODO

*   Implement Enemy Snakes (AI, collisions, rendering).
*   Integrate Sound Effects and Background Music.
*   Add Visual Effects (particles, speed indicators).
*   Improve Graphics (use sprites/textures).
*   Implement UI (Main Menu, Pause Menu, HUD).

