package main

import (
	"errors"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	colorLight color.Color = color.RGBA{199, 240, 216, 255}
	colorDark  color.Color = color.RGBA{67, 82, 61, 255}
)

func main() {
	ebiten.SetWindowSize(840, 480)
	ebiten.SetWindowTitle("TODO: cool game name")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	gameWidth, gameHeight := 84, 48

	game := &Game{
		Width:  gameWidth,
		Height: gameHeight,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	Width  int
	Height int
}

// Update updates a game by one tick. The given argument represents a screen image.
//
// Update updates only the game logic and Draw draws the screen.
//
// In the first frame, it is ensured that Update is called at least once before Draw. You can use Update
// to initialize the game state.
//
// After the first frame, Update might not be called or might be called once
// or more for one frame. The frequency is determined by the current TPS (tick-per-second).
func (g *Game) Update() error {
	// Pressing Esc any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	return nil
}

// Draw draws the game screen by one frame.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorDark)
	ebitenutil.DrawLine(screen, 0, 0, 47, 47, color.RGBA{199, 240, 216, 255})
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}
