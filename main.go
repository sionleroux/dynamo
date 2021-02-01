package main

import (
	"errors"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
		Player: &Player{},
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	Width  int
	Height int
	Player *Player
}

// Update updates a game by one tick. The given argument represents a screen image.
func (g *Game) Update() error {
	// Pressing Esc any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	// Movement controls
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if g.Player.y+1 <= float64(g.Height-1) {
			g.Player.y++
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if g.Player.y-1 >= 0 {
			g.Player.y--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		if g.Player.x-1 >= 0 {
			g.Player.x--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		if g.Player.x+1 <= float64(g.Width-1) {
			g.Player.x++
		}
	}

	return nil
}

// Player is the pixel the player controlers
type Player struct {
	x float64
	y float64
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorDark)
	ebitenutil.DrawRect(screen, g.Player.x, g.Player.y, 1, 1, color.RGBA{199, 240, 216, 255})
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}
