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
}

// Update updates a game by one tick.
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
	if maze.Image.At(newCoords.X, newCoords.Y) != ColorDark {
		p.Coords = newCoords
		p.Step = 2 // short cooldown when holding down
		if inpututil.IsKeyJustPressed(key) {
			p.Step = 15 // long first cooldown when tapping key
		}
	}
}

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
	colorMaze := image.NewPaletted(grayMaze.Bounds(), NokiaPalette)
	for k, v := range grayMaze.Pix {
		if v == 255 {
			colorMaze.Pix[k] = 1
		}
	}

	// Find an exit at the bottom right
	var exit image.Point
	for i := colorMaze.Rect.Max.X; i > 0; i-- {
		if colorMaze.At(i, colorMaze.Rect.Max.Y-2) == ColorLight {
			exit = image.Pt(i, colorMaze.Rect.Max.Y)
			colorMaze.Set(exit.X, exit.Y-1, ColorLight)
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
