package server

import "mmos/internal/common/geom"

// WindowInput is input addressed to a window in that window's local space.
type WindowInput struct {
	WindowID WindowID
	Event    InputEvent
}

// InputDispatcher applies display-wide input policy to normalized device
// events. Platform-specific readers turn hardware input into InputEvent before
// it reaches the dispatcher.
//
// A pointer stays with the window chosen on PointerDown until it ends or is
// cancelled. This prevents a drag from changing target as it crosses another
// visible window.
type InputDispatcher struct {
	capture map[int]WindowID
}

func NewInputDispatcher() InputDispatcher {
	return InputDispatcher{capture: map[int]WindowID{}}
}

// Dispatch chooses an input target and converts coordinate-bearing input to
// the target window's local space. Keyboard and text input target the focused
// window. Scroll input targets the window below its position without changing
// keyboard focus.
func (d *InputDispatcher) Dispatch(event InputEvent, windows *WindowServer) (WindowInput, bool) {
	switch event := event.(type) {
	case PointerEvent:
		return d.dispatchPointer(event, windows)
	case ScrollEvent:
		id, ok := windows.WindowAt(event.X, event.Y)
		return localWindowInput(id, event, windows, ok)
	}
	id, ok := windows.FocusedWindowID()
	return localWindowInput(id, event, windows, ok)
}

func (d *InputDispatcher) dispatchPointer(event PointerEvent, windows *WindowServer) (WindowInput, bool) {
	id, captured := d.capture[event.ChangedPointerID]
	if !captured {
		if event.Action != PointerDown {
			return WindowInput{}, false
		}
		pointer, ok := event.PointerByID(event.ChangedPointerID)
		if !ok {
			return WindowInput{}, false
		}
		id, ok = windows.FocusAt(pointer.X, pointer.Y)
		if !ok {
			return WindowInput{}, false
		}
		d.capture[event.ChangedPointerID] = id
	}

	routed, ok := localWindowInput(id, event, windows, true)
	if event.Action == PointerUp || event.Action == PointerCancel || !ok {
		delete(d.capture, event.ChangedPointerID)
	}
	return routed, ok
}

func (d *InputDispatcher) Release(id WindowID) {
	for pointerID, captured := range d.capture {
		if captured == id {
			delete(d.capture, pointerID)
		}
	}
}

func localWindowInput(id WindowID, event InputEvent, windows *WindowServer, ok bool) (WindowInput, bool) {
	if !ok {
		return WindowInput{}, false
	}
	window := windows.Windows[id]
	if window == nil || !window.Visible {
		return WindowInput{}, false
	}
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
