package server

import "mmos/internal/common/geom"

// DisplayID identifies one logical display in the system.
type DisplayID string

const PrimaryDisplay DisplayID = "primary"

// DisplayConfig is the logical display state adopted for a composed frame.
type DisplayConfig struct {
	Width, Height int
	Revision      int
	// SafeArea is reserved for trusted chrome. Apps may use ContentBounds to
	// avoid it instead of guessing a status-bar size.
	SafeArea geom.Insets
	// Cutout identifies the physical exclusion, in display coordinates. It is
	// informational when the same area is already included in SafeArea.
	Cutout  geom.Rect
	Density float32
}

func (c DisplayConfig) Bounds() geom.Rect {
	return geom.Rect{Width: c.Width, Height: c.Height}
}

// ContentBounds is the application-safe logical area of the display.
func (c DisplayConfig) ContentBounds() geom.Rect { return c.Bounds().Inset(c.SafeArea) }
