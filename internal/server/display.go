package server

import "mmos/internal/common/geom"

// DisplayID identifies one logical display in the system.
type DisplayID string

const PrimaryDisplay DisplayID = "primary"

type Orientation uint8

const (
	TopUp Orientation = iota
	RightUp
	BottomUp
	LeftUp
)

type ScreenPower uint8

const (
	ScreenOn ScreenPower = iota
	ScreenOff
)

// DisplayConfig is the logical display state adopted for a composed frame.
type DisplayConfig struct {
	Width, Height int
	Revision      int
	// SafeArea is reserved for trusted chrome. Apps may use ContentBounds to
	// avoid it instead of guessing a status-bar size.
	SafeArea geom.Insets
	// Cutout identifies the physical exclusion, in display coordinates. It is
	// informational when the same area is already included in SafeArea.
	Cutout      geom.Rect
	Density     float32
	Orientation Orientation
	Power       ScreenPower
}

func (c DisplayConfig) Bounds() geom.Rect {
	return geom.Rect{Width: c.Width, Height: c.Height}
}

// ContentBounds is the application-safe logical area of the display.
func (c DisplayConfig) ContentBounds() geom.Rect { return c.Bounds().Inset(c.SafeArea) }

func (c DisplayConfig) Rotate() DisplayConfig {
	oldWidth := c.Width
	cutout := c.Cutout
	c.Width, c.Height = c.Height, c.Width
	c.Orientation = (c.Orientation + 1) % 4
	c.SafeArea = geom.Insets{Top: c.SafeArea.Left, Right: c.SafeArea.Top, Bottom: c.SafeArea.Right, Left: c.SafeArea.Bottom}
	c.Cutout = geom.Rect{X: oldWidth - cutout.Y - cutout.Height, Y: cutout.X, Width: cutout.Height, Height: cutout.Width}
	c.Revision++
	return c
}
