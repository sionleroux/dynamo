package main

import (
	"image"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sinisterstuf/dynamo/media"
	"gitlab.com/zaba505/maze"
)

// Maze contains all information about mazes
// Not just the generated maze image but also any other meta-data that can be
// used for interacting with the maze.
type Maze struct {
	Image  *ebiten.Image // Maze image in 1-bit for drawing & collision logic
	Maze   *maze.Maze    // Original maze object for solving
	Exit   image.Point   // The exit location, for end-game logic
	Offset image.Point   // Used to centre the maze at draw time
}

// NewMaze generates a new maze based on difficulty level and random source
func NewMaze(source rand.Source, level int, gameSize image.Point) *Maze {
	mymaze := maze.WithKruskal(source).Generate(
		gameSize.X/Levels[level]-1,
		gameSize.Y/Levels[level]-1,
	)

	// Convert to Nokia colours
	grayMaze := maze.Gray(mymaze)
	colorMaze := image.NewPaletted(grayMaze.Bounds(), media.NokiaPalette)
	for k, v := range grayMaze.Pix {
		if v == 255 {
			colorMaze.Pix[k] = 1
		}
	}

	// Find an exit at the bottom right
	var exit image.Point
	for i := colorMaze.Rect.Max.X; i > 0; i-- {
		if colorMaze.At(i, colorMaze.Rect.Max.Y-2) == media.ColorLight {
			exit = image.Pt(i, colorMaze.Rect.Max.Y)
			colorMaze.Set(exit.X, exit.Y-1, media.ColorLight)
			break
		}
	}
	mazeImage := ebiten.NewImageFromImage(colorMaze)

	// Calculate offset from origin for centring on the screen
	offset := image.Pt(
		(gameSize.X-mazeImage.Bounds().Dx())/2,
		(gameSize.Y-mazeImage.Bounds().Dy())/2,
	)

	return &Maze{
		Maze:   mymaze,
		Image:  mazeImage,
		Exit:   exit,
		Offset: offset,
	}
}
