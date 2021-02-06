package main

import (
	"errors"
	"image"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"gitlab.com/zaba505/maze"
)

var (
	colorLight   color.Color   = color.RGBA{199, 240, 216, 255}
	colorDark    color.Color   = color.RGBA{67, 82, 61, 255}
	nokiaPalette color.Palette = color.Palette{colorDark, colorLight}
)

func main() {
	ebiten.SetWindowSize(840, 480)
	ebiten.SetWindowTitle("Dynamo")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	ebiten.SetWindowResizable(true)
	gameWidth, gameHeight := 84, 48

	mymaze := maze.WithKruskal(rand.NewSource(1)).Generate(gameWidth/2-1, gameHeight/2-1)
	grayMaze := maze.Gray(mymaze)
	colorMaze := image.NewPaletted(grayMaze.Bounds(), nokiaPalette)
	for k, v := range grayMaze.Pix {
		if v == 255 {
			colorMaze.Pix[k] = 1
		}
	}

	var exit image.Point
	for i := colorMaze.Rect.Max.X; i > 0; i-- {
		if colorMaze.At(i, colorMaze.Rect.Max.Y-2) == colorLight {
			exit = image.Pt(i, colorMaze.Rect.Max.Y)
			colorMaze.Set(exit.X, exit.Y-1, colorLight)
			break
		}
	}

	mazeImage := ebiten.NewImageFromImage(colorMaze)

	game := &Game{
		Width:   gameWidth,
		Height:  gameHeight,
		Player:  &Player{image.Pt(1, 1), true, false},
		Maze:    mazeImage,
		Exit:    exit,
		BlinkOn: true,
	}

	blinker := time.NewTicker(500 * time.Millisecond)
	mover := time.NewTicker(150 * time.Millisecond)

	go func() {
		for {
			select {
			case <-blinker.C:
				game.BlinkOn = !game.BlinkOn
			case <-mover.C:
				game.Player.Moving = false
			}
		}
	}()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game tracks global game states
type Game struct {
	Width   int
	Height  int
	Player  *Player
	Maze    *ebiten.Image
	Exit    image.Point
	BlinkOn bool
}

// Update updates a game by one tick. The given argument represents a screen image.
func (g *Game) Update() error {
	// Pressing Esc any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	if g.Player.Coords.Eq(g.Exit) {
		return errors.New("you win")
	}

	// Movement controls
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.Player.Move(g.Maze, image.Pt(0, 1))
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.Player.Move(g.Maze, image.Pt(0, -1))
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.Player.Move(g.Maze, image.Pt(-1, 0))
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.Player.Move(g.Maze, image.Pt(1, 0))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.Player.TorchOn = !g.Player.TorchOn
	}

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorDark)
	if g.Player.TorchOn {
		screen.DrawImage(g.Maze, &ebiten.DrawImageOptions{})
	}
	playercolor := colorDark
	if g.BlinkOn || !g.Player.TorchOn {
		playercolor = colorLight
	}
	ebitenutil.DrawRect(screen, float64(g.Player.Coords.X), float64(g.Player.Coords.Y), 1, 1, playercolor)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

// Player is the pixel the player controlers
type Player struct {
	Coords  image.Point
	TorchOn bool
	Moving  bool
}

func (p *Player) Move(maze *ebiten.Image, dest image.Point) {
	if p.Moving {
		return
	}

	p.TorchOn = false
	newCoords := p.Coords.Add(dest)
	if maze.At(newCoords.X, newCoords.Y) != colorDark {
		p.Coords = newCoords
		p.Moving = true
	}
}
