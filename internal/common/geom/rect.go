// Package geom defines shared geometric values.
package geom

type Rect struct {
	X, Y          int
	Width, Height int
}

// Insets reserves display edges for system UI or a physical cutout.
type Insets struct {
	Top, Right, Bottom, Left int
}

// Inset returns the part of r that remains usable after reserving edges.
func (r Rect) Inset(insets Insets) Rect {
	width := r.Width - insets.Left - insets.Right
	height := r.Height - insets.Top - insets.Bottom
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return Rect{X: r.X + insets.Left, Y: r.Y + insets.Top, Width: width, Height: height}
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}
