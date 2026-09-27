package server

import "mmos/internal/common/geom"

// WindowInput is input addressed to a window in that window's local space.
type WindowInput struct {
	WindowID WindowID
	Event    InputEvent
}

// InputServer selects the foreground application window and normalizes input
// coordinates before the application receives the event.
type InputServer struct{}

func NewInputServer() InputServer { return InputServer{} }

func (InputServer) Route(event InputEvent, windows *WindowServer) (WindowInput, bool) {
	id, ok := windows.FocusedWindowID()
	if !ok {
		return WindowInput{}, false
	}
	window := windows.Windows[id]
	return WindowInput{WindowID: id, Event: localInput(event, window.Bounds)}, true
}

func localInput(event InputEvent, bounds geom.Rect) InputEvent {
	switch event := event.(type) {
	case PointerEvent:
		pointers := make([]PointerSample, len(event.Pointers))
		for i, pointer := range event.Pointers {
			pointer.X -= bounds.X
			pointer.Y -= bounds.Y
			pointers[i] = pointer
		}
		event.Pointers = pointers
		return event
	case ScrollEvent:
		event.X -= bounds.X
		event.Y -= bounds.Y
		return event
	default:
		return event
	}
}
