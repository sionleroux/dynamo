package main

import (
	"errors"
	"image"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"gitlab.com/zaba505/maze"
)

var (
	colorLight   color.Color   = color.RGBA{199, 240, 216, 255}
	colorDark    color.Color   = color.RGBA{67, 82, 61, 255}
	nokiaPalette color.Palette = color.Palette{
		colorDark,
		colorLight,
	}
)

func main() {
	ebiten.SetWindowSize(840, 480)
	ebiten.SetWindowTitle("Dynamo")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	ebiten.SetWindowResizable(true)
	gameWidth, gameHeight := 84, 48

	source := rand.NewSource(1)
	generator := maze.WithKruskal(source)
	mymaze := generator.Generate(gameWidth/2-1, gameHeight/2-1)
	grayMaze := maze.Gray(mymaze)
	colorMaze := image.NewPaletted(grayMaze.Bounds(), nokiaPalette)
	for k, v := range grayMaze.Pix {
		if v == 255 {
			colorMaze.Pix[k] = 1
		}
	}
	mazeImage := ebiten.NewImageFromImage(colorMaze)

	game := &Game{
		Width:  gameWidth,
		Height: gameHeight,
		Player: &Player{1, 1, true},
		Maze:   mazeImage,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game tracks global game states
type Game struct {
	Width  int
	Height int
	Player *Player
	Maze   *ebiten.Image
}

// Update updates a game by one tick. The given argument represents a screen image.
func (g *Game) Update() error {
	// Pressing Esc any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	// Movement controls
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if g.Player.Y+1 <= float64(g.Height-1) {
			g.Player.Y++
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if g.Player.Y-1 >= 0 {
			g.Player.Y--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		if g.Player.X-1 >= 0 {
			g.Player.X--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		if g.Player.X+1 <= float64(g.Width-1) {
			g.Player.X++
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		if g.Player.TorchOn {
			g.Player.TorchOn = false
		} else {
			g.Player.TorchOn = true
		}
	}

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorDark)
	if g.Player.TorchOn {
		screen.DrawImage(g.Maze, &ebiten.DrawImageOptions{})
	}
	ebitenutil.DrawRect(screen, g.Player.X, g.Player.Y, 1, 1, color.RGBA{199, 240, 216, 255})
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

// Player is the pixel the player controlers
type Player struct {
	X       float64
	Y       float64
	TorchOn bool
}
