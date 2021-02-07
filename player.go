package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/sinisterstuf/dynamo/media"
)

// Player is the pixel the player controls
type Player struct {
	Coords  image.Point
	TorchOn bool
	Step    int
	Moved   bool
}

// NewPlayer initialises a new Player object with default values
func NewPlayer() *Player {
	return &Player{
		Coords:  image.Pt(1, 1), // This is inset by 1 because 0,0 is a wall
		TorchOn: true,           // Start with torch on so that the map is shown
	}
}

// Move moves the Player in the given direction if possible
// This includes checks for whether the move is legal at all, e.g. would collide
// with a wall.  It also includes special logic for movement speed when the key
// is being tapped and when it's being held down.
func (p *Player) Move(maze *Maze, dest image.Point, key ebiten.Key) {

	// Still cooling down from last move, unless the key was tapped
	if p.Step > 0 && !inpututil.IsKeyJustPressed(key) {
		return
	}

	// Don't move if the key is still being held in from the last level
	if !p.Moved {
		if inpututil.IsKeyJustPressed(key) {
			p.Moved = true
		} else { // Skip return so that lower-down "just pressed" logic runs
			return
		}
	}

	// Even just attempting to move turns off the torch
	p.TorchOn = false

	// Do the actual move if legal
	newCoords := p.Coords.Add(dest)
	if maze.Image.At(newCoords.X, newCoords.Y) != media.ColorDark {
		p.Coords = newCoords
		p.Step = 2 // short cooldown when holding down
		if inpututil.IsKeyJustPressed(key) {
			p.Step = 15 // long first cooldown when tapping key
		}
	}
}
