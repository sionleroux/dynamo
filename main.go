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
	Levels       []int         = []int{10, 6, 4, 3, 2}
)

const (
	LevelBeginner int = iota
	LevelEasy
	LevelMedium
	LevelHard
	LevelExtreme
)

func main() {
	gameSize := image.Pt(84, 48)
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
	}

	go func() {
		blinker := time.NewTicker(500 * time.Millisecond)
		for _ = range blinker.C {
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
}

// Update updates a game by one tick. The given argument represents a screen image.
func (g *Game) Update() error {
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
	screen.Fill(colorDark)
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
			colorLight,
		)
	}
	playercolor := colorDark
	if g.BlinkOn || !g.Player.TorchOn {
		playercolor = colorLight
	}
	playerPos := g.Player.Coords.Add(g.Maze.Offset)
	screen.Set(playerPos.X, playerPos.Y, playercolor)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Size.X, g.Size.Y
}

func (g *Game) NextLevel() {
	g.Win = false
	g.Player = NewPlayer()
	if g.Level <= LevelExtreme {
		g.Level++
	}
	g.Maze = NewMaze(g.Source, g.Level, g.Size)
}

// Player is the pixel the player controlers
type Player struct {
	Coords  image.Point
	TorchOn bool
	Step    int
	Moved   bool
}

func NewPlayer() *Player {
	return &Player{
		Coords:  image.Pt(1, 1),
		TorchOn: true,
	}
}

func (p *Player) Move(maze *Maze, dest image.Point, key ebiten.Key) {
	if p.Step > 0 && !inpututil.IsKeyJustPressed(key) {
		return
	}

	// Don't move if the key is still being held in from the last level
	if !p.Moved {
		if inpututil.IsKeyJustPressed(key) {
			p.Moved = true
		}
		return
	}

	p.TorchOn = false
	newCoords := p.Coords.Add(dest)
	if maze.Image.At(newCoords.X, newCoords.Y) != colorDark {
		p.Coords = newCoords
		p.Step = 2
		if inpututil.IsKeyJustPressed(key) {
			p.Step = 15
		}
	}
}

type Maze struct {
	Image  *ebiten.Image
	Maze   *maze.Maze
	Exit   image.Point
	Offset image.Point
}

func NewMaze(source rand.Source, level int, gameSize image.Point) *Maze {
	mymaze := maze.WithKruskal(source).Generate(
		gameSize.X/Levels[level]-1,
		gameSize.Y/Levels[level]-1,
	)

	// Convert to Nokia colours
	grayMaze := maze.Gray(mymaze)
	colorMaze := image.NewPaletted(grayMaze.Bounds(), nokiaPalette)
	for k, v := range grayMaze.Pix {
		if v == 255 {
			colorMaze.Pix[k] = 1
		}
	}

	// Find an exit at the bottom right
	var exit image.Point
	for i := colorMaze.Rect.Max.X; i > 0; i-- {
		if colorMaze.At(i, colorMaze.Rect.Max.Y-2) == colorLight {
			exit = image.Pt(i, colorMaze.Rect.Max.Y)
			colorMaze.Set(exit.X, exit.Y-1, colorLight)
			break
		}
	}

	mazeImage := ebiten.NewImageFromImage(colorMaze)

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
