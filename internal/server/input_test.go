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

func TestInputServerRoutesForegroundInputInWindowCoordinates(t *testing.T) {
	windows := NewWindowServer()
	windows.Register("chat", WindowConfig{
		Bounds: geom.Rect{X: 100, Y: 200, Width: 320, Height: 480},
	})
	if !windows.Activate("chat") {
		t.Fatal("activate chat")
	}
	input := NewInputServer()
	routed, ok := input.Route(PointerEvent{
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

	for _, event := range []InputEvent{} {
		if _, ok := input.Route(event, &windows); ok {
			t.Fatalf("%T must not route to an app window", event)
		}
	}
}
