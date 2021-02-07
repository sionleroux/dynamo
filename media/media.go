// Package media provides graphics for the game generated from PNG files
package media

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	// ColorLight is the ON or 1 screen colour, similar to white
	ColorLight color.Color = color.RGBA{199, 240, 216, 255}
	// ColorDark is the OFF or 0 screen colour, similar to black
	ColorDark color.Color = color.RGBA{67, 82, 61, 255}
	// NokiaPalette is a 1-bit palette of greenish colours simulating Nokia 3310
	NokiaPalette color.Palette = color.Palette{ColorDark, ColorLight}
	// GameSize is the screen resolution of a Nokia 3310
	GameSize image.Point = image.Point{84, 48}
)

// Animation is a set of frames that can be stepped and drawn
type Animation struct {
	Frames     []*ebiten.Image
	Index      int
	Delay      int
	delayCount int
}

// CurrentFrame returns an ebiten Image for the current frame
func (a *Animation) CurrentFrame() *ebiten.Image {
	return a.Frames[a.Index]
}

// Update steps through frames with a delay
func (a *Animation) Update() {
	if a.delayCount == 0 {
		a.nextFrame()
	}
	a.delayCount = (a.delayCount + 1) % a.Delay
}

// steps through frames
func (a *Animation) nextFrame() {
	a.Index = (a.Index + 1) % (len(a.Frames) - 1)
}

// NewTitleFrames generates an animation of title frames
func NewTitleFrames() *Animation {
	frames := make([]*ebiten.Image, 6)

	for k, v := range [][]uint8{
		Title_waiting_1,
		Title_waiting_2,
		Title_waiting_3,
		Title_waiting_4,
		Title_waiting_5,
		Title_waiting_6,
	} {
		frame := image.NewPaletted(
			image.Rectangle{image.Point{}, GameSize},
			NokiaPalette,
		)
		frame.Pix = v
		frames[k] = ebiten.NewImageFromImage(frame)
	}

	return &Animation{
		Frames: frames,
		Delay:  10,
	}
}
