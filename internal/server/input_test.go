package server

import (
	"testing"

	"mmos/internal/common/geom"
)

func TestInputEventsImplementInputEvent(t *testing.T) {
	pointer := PointerEvent{
		Source:           TouchSource,
		Action:           PointerMove,
		ChangedPointerID: 2,
		Pointers: []PointerSample{
			{ID: 0, X: 10, Y: 20},
			{ID: 2, X: 30, Y: 40},
		},
	}
	changed, ok := pointer.PointerByID(2)
	if !ok || changed.X != 30 || changed.Y != 40 {
		t.Fatalf("changed pointer = %#v, found=%t", changed, ok)
	}

	for _, event := range []InputEvent{
		pointer,
		KeyEvent{Action: KeyDown, Key: "VolumeUp"},
		TextEvent{Text: "안녕하세요"},
		ScrollEvent{Source: StylusSource, X: 10, Y: 20, DeltaY: -24},
	} {
		if event == nil {
			t.Fatal("normalized input event must not be nil")
		}
	}
}

func TestInputDispatcherRoutesCapturedPointerInWindowCoordinates(t *testing.T) {
	windows := NewWindowServer()
	windows.Register("chat", WindowConfig{
		Bounds: geom.Rect{X: 100, Y: 200, Width: 320, Height: 480},
	})
	if !windows.Activate("chat") {
		t.Fatal("activate chat")
	}
	input := NewInputDispatcher()
	if _, ok := input.Dispatch(PointerEvent{
		Action:           PointerDown,
		Source:           TouchSource,
		ChangedPointerID: 0,
		Pointers:         []PointerSample{{ID: 0, X: 160, Y: 236}},
	}, &windows); !ok {
		t.Fatal("dispatch pointer down")
	}
	routed, ok := input.Dispatch(PointerEvent{
		Action:           PointerUp,
		Source:           TouchSource,
		ChangedPointerID: 0,
		Pointers:         []PointerSample{{ID: 0, X: 160, Y: 236}},
	}, &windows)
	if !ok || routed.WindowID != "chat" {
		t.Fatalf("route = %#v, found=%t", routed, ok)
	}
	pointer := routed.Event.(PointerEvent)
	if point, ok := pointer.PointerByID(0); !ok || point.X != 60 || point.Y != 36 {
		t.Fatalf("window-local pointer = %#v, found=%t", point, ok)
	}

	if _, ok := input.Dispatch(PointerEvent{Action: PointerUp, ChangedPointerID: 1, Pointers: []PointerSample{{ID: 1}}}, &windows); ok {
		t.Fatal("pointer without a preceding down must not route")
	}
}

func TestInputDispatcherKeepsPointerCaptureAcrossWindows(t *testing.T) {
	windows := NewWindowServer()
	windows.Register("left", WindowConfig{Bounds: geom.Rect{Width: 160, Height: 480}})
	windows.Register("right", WindowConfig{Bounds: geom.Rect{X: 160, Width: 160, Height: 480}})
	if !windows.Activate("left") || !windows.Activate("right") {
		t.Fatal("show both windows")
	}

	input := NewInputDispatcher()
	if _, ok := input.Dispatch(PointerEvent{
		Action: PointerDown, ChangedPointerID: 0,
		Pointers: []PointerSample{{ID: 0, X: 40, Y: 20}},
	}, &windows); !ok {
		t.Fatal("dispatch pointer down")
	}
	routed, ok := input.Dispatch(PointerEvent{
		Action: PointerMove, ChangedPointerID: 0,
		Pointers: []PointerSample{{ID: 0, X: 200, Y: 20}},
	}, &windows)
	if !ok || routed.WindowID != "left" || windows.Foreground != "left" {
		t.Fatalf("route = %#v foreground=%q", routed, windows.Foreground)
	}
	point, ok := routed.Event.(PointerEvent).PointerByID(0)
	if !ok || point.X != 200 || point.Y != 20 {
		t.Fatalf("local point = %#v found=%t", point, ok)
	}
}

func TestInputDispatcherRejectsPointerDownOutsideVisibleWindows(t *testing.T) {
	windows := NewWindowServer()
	windows.Register("chat", WindowConfig{Bounds: geom.Rect{Width: 160, Height: 480}})
	if !windows.Activate("chat") {
		t.Fatal("activate chat")
	}
	input := NewInputDispatcher()
	if _, ok := input.Dispatch(PointerEvent{
		Action: PointerDown, ChangedPointerID: 0,
		Pointers: []PointerSample{{ID: 0, X: 200, Y: 20}},
	}, &windows); ok {
		t.Fatal("pointer outside every visible window must not route")
	}
}

func TestInputDispatcherScrollDoesNotChangeFocus(t *testing.T) {
	windows := NewWindowServer()
	windows.Register("left", WindowConfig{Bounds: geom.Rect{Width: 160, Height: 480}})
	windows.Register("right", WindowConfig{Bounds: geom.Rect{X: 160, Width: 160, Height: 480}})
	if !windows.Activate("left") || !windows.Activate("right") {
		t.Fatal("prepare focused left window")
	}
	if _, ok := windows.FocusAt(20, 20); !ok {
		t.Fatal("focus left window")
	}

	input := NewInputDispatcher()
	routed, ok := input.Dispatch(ScrollEvent{X: 200, Y: 20, DeltaY: 10}, &windows)
	if !ok || routed.WindowID != "right" || windows.Foreground != "left" {
		t.Fatalf("route = %#v foreground=%q", routed, windows.Foreground)
	}
}
