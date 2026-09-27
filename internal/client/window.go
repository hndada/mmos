package client

import (
	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

// Window is owned by the client process. It contains the UI tree and the
// buffer that the client submits to the model.
type Window struct {
	ID     model.WindowID
	Root   *UINode
	Buffer model.Buffer
}

func (w *Window) CommandAt(event model.InputEvent) (Command, bool) {
	// A command is activated only when a primary pointing device is released.
	// Other normalized inputs remain available to application UI code but do
	// not accidentally turn into a button command during hit testing.
	pointerEvent, ok := event.(model.PointerEvent)
	if !ok || pointerEvent.Action != model.PointerUp ||
		pointerEvent.ChangedPointerID != 0 {
		return "", false
	}
	pointer, ok := pointerEvent.PointerByID(pointerEvent.ChangedPointerID)
	if !ok {
		return "", false
	}
	node := w.HitTest(pointer.X, pointer.Y)
	if node == nil || !node.Enabled || node.Command == "" {
		return "", false
	}
	return node.Command, true
}

func (w *Window) HitTest(x, y int) *UINode {
	return w.Root.HitTest(x, y)
}

func (w *Window) Node(id string) *UINode {
	return w.Root.Find(id)
}

// Bounds returns the root UI extent in its display-relative logical space.
func (w *Window) Bounds() geom.Rect {
	if w.Root == nil {
		return geom.Rect{}
	}
	return geom.Rect{Width: w.Root.Bounds.Width, Height: w.Root.Bounds.Height}
}

// Resize updates the client layout extent after the server changes display
// configuration. The next Drawer call produces a buffer for these bounds.
func (w *Window) Resize(bounds geom.Rect) {
	if w.Root == nil {
		return
	}
	w.Root.Bounds = bounds
	w.Buffer = model.Buffer{}
}

func NewWindow(id model.WindowID, rootID string) *Window {
	return &Window{
		ID: id,
		Root: &UINode{
			ID: rootID, Kind: Container,
			Bounds: geom.Rect{Width: 320, Height: 480}, Visible: true,
		},
	}
}
