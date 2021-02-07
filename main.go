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
	"github.com/sinisterstuf/dynamo/media"
)

var (
	// ColorLight is the ON or 1 screen colour, similar to white
	ColorLight color.Color = color.RGBA{199, 240, 216, 255}
	// ColorDark is the OFF or 0 screen colour, similar to black
	ColorDark color.Color = color.RGBA{67, 82, 61, 255}
	// NokiaPalette is a 1-bit palette of greenish colours simulating Nokia 3310
	NokiaPalette color.Palette = color.Palette{ColorDark, ColorLight}
	// Levels maps level difficulty indices to maze size scaling factors
	Levels []int = []int{10, 6, 4, 3, 2}
)

// Levels represent the difficulty of different game levels
const (
	LevelBeginner int = iota
	LevelEasy
	LevelMedium
	LevelHard
	LevelExtreme
)

// State is a high-level game state controlling app behaviour
type State int

// States represent different overall states the game can be in
const (
	StateTitle State = iota
	StateMenu
	StateLevel
)

func main() {
	gameSize := image.Pt(84, 48)
	windowScale := 10
	ebiten.SetWindowSize(gameSize.X*windowScale, gameSize.Y*windowScale)
	ebiten.SetWindowTitle("Dynamo")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	ebiten.SetWindowResizable(true)

	source := rand.NewSource(int64(time.Now().Nanosecond()))

	title := image.NewPaletted(
		image.Rectangle{image.Point{}, gameSize},
		NokiaPalette,
	)
	title.Pix = media.Title

	game := &Game{
		Size:    gameSize,
		Player:  NewPlayer(),
		Maze:    NewMaze(source, LevelBeginner, gameSize),
		BlinkOn: true,
		Win:     false,
		Level:   LevelBeginner,
		Source:  source,
		Title:   ebiten.NewImageFromImage(title),
	}

	go func() {
		blinker := time.NewTicker(500 * time.Millisecond)
		for range blinker.C {
			game.BlinkOn = !game.BlinkOn
		}
	}()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game tracks global game states
type Game struct {
	Size    image.Point
	Player  *Player
	Maze    *Maze
	BlinkOn bool
	Win     bool
	Level   int
	Source  rand.Source
	State   State
	Title   *ebiten.Image
}

// Update updates a game by one tick.
func (g *Game) Update() error {
	switch g.State {
	case StateTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeyE) {
			g.State = StateLevel
		}
	case StateLevel:
		return updateLevel(g)
	}
	return nil
}

func updateLevel(g *Game) error {
	// Pressing Esc any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	if g.Player.Coords.Eq(g.Maze.Exit) {
		g.Win = true
	}

	if g.Win {
		g.Player.Coords.Y++
		if g.Player.Coords.Y > g.Size.Y {
			g.NextLevel()
		}
		return nil
	}

	// Movement controls
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.Player.Move(g.Maze, image.Pt(0, 1), ebiten.KeyS)
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.Player.Move(g.Maze, image.Pt(0, -1), ebiten.KeyW)
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.Player.Move(g.Maze, image.Pt(-1, 0), ebiten.KeyA)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.Player.Move(g.Maze, image.Pt(1, 0), ebiten.KeyD)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.Player.TorchOn = !g.Player.TorchOn
	}

	if g.Player.Step > 0 {
		g.Player.Step--
	}

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.State {
	case StateTitle:
		screen.DrawImage(g.Title, &ebiten.DrawImageOptions{})
	case StateLevel:
		drawLevel(g, screen)
	}
}

func drawLevel(g *Game, screen *ebiten.Image) {
	screen.Fill(ColorDark)
	if g.Player.TorchOn {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(g.Maze.Offset.X),
			float64(g.Maze.Offset.Y),
		)
		screen.DrawImage(g.Maze.Image, op)
		ebitenutil.DrawLine(
			screen,
			float64(g.Maze.Exit.X+g.Maze.Offset.X+1),
			float64(g.Maze.Exit.Y+g.Maze.Offset.Y),
			float64(g.Maze.Exit.X+g.Maze.Offset.X+1),
			float64(screen.Bounds().Max.Y),
			ColorLight,
		)
	}
	playercolor := ColorDark
	if g.BlinkOn || !g.Player.TorchOn {
		playercolor = ColorLight
	}
	playerPos := g.Player.Coords.Add(g.Maze.Offset)
	screen.Set(playerPos.X, playerPos.Y, playercolor)
}

// Layout scales the pixels when the windows is resized
// This means that in a bigger window all the pixels will become bigger squares
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Size.X, g.Size.Y
}

// NextLevel sets up the next level of the game
// It handles things like increasing difficulty and resetting the Player state
func (g *Game) NextLevel() {
	g.Win = false
	g.Player = NewPlayer()
	if g.Level <= LevelExtreme {
		g.Level++
	}
	g.Maze = NewMaze(g.Source, g.Level, g.Size)
}
