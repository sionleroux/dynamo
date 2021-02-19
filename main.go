package main

import (
	"errors"
	"image"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/sinisterstuf/dynamo/media"
)

// Levels maps level difficulty indices to maze size scaling factors
var Levels []int = []int{10, 6, 4, 3, 2}

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
	StateTitleTransition
	StateMenu
	StateLevel
)

func main() {
	gameSize := media.GameSize
	windowScale := 10
	ebiten.SetWindowSize(gameSize.X*windowScale, gameSize.Y*windowScale)
	ebiten.SetWindowTitle("Dynamo")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	ebiten.SetWindowResizable(true)

	source := rand.NewSource(int64(time.Now().Nanosecond()))

	game := &Game{
		Size:    gameSize,
		Player:  NewPlayer(),
		Maze:    NewMaze(source, LevelBeginner, gameSize),
		BlinkOn: true,
		Win:     false,
		Level:   LevelBeginner,
		Source:  source,
		Title:   media.NewTitleFrames(),
		TT:      media.NewTitleTransitionFrames(),
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
	Title   *media.Animation
	TT      *media.Animation
}

// Update updates a game by one tick.
func (g *Game) Update() error {
	switch g.State {
	case StateTitle:
		g.Title.Update()
		if inpututil.IsKeyJustPressed(ebiten.KeyE) {
			g.State = StateTitleTransition
		}
	case StateTitleTransition:
		g.TT.Update()
		if g.TT.Index == 0 {
			g.State = StateLevel
		}
	case StateLevel:
		return updateLevel(g)
	}
	return nil
}

func updateLevel(g *Game) error {
	// Pressing Q quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
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
		screen.DrawImage(g.Title.CurrentFrame(), &ebiten.DrawImageOptions{})
	case StateTitleTransition:
		screen.DrawImage(g.TT.CurrentFrame(), &ebiten.DrawImageOptions{})
	case StateLevel:
		drawLevel(g, screen)
	}
}

func drawLevel(g *Game, screen *ebiten.Image) {
	screen.Fill(media.ColorDark)
	if g.Player.TorchOn {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(g.Maze.Offset.X),
			float64(g.Maze.Offset.Y),
		)
		// torchLight := image.NewPaletted(g.Maze.Image.Bounds(), media.NokiaPalette)

		screen.DrawImage(g.Maze.Image, op)
		ebitenutil.DrawLine(
			screen,
			float64(g.Maze.Exit.X+g.Maze.Offset.X+1),
			float64(g.Maze.Exit.Y+g.Maze.Offset.Y),
			float64(g.Maze.Exit.X+g.Maze.Offset.X+1),
			float64(screen.Bounds().Max.Y),
			media.ColorLight,
		)
	}
	playercolor := media.ColorDark
	if g.BlinkOn || !g.Player.TorchOn {
		playercolor = media.ColorLight
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
	if g.Level < LevelExtreme {
		g.Level++
	}
	g.Maze = NewMaze(g.Source, g.Level, g.Size)
}
