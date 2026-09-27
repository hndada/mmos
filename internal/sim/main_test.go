package sim

import (
	"testing"

	"mmos/internal/server"
)

func TestSimulatorCoreFlow(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, changed := s.Input(tap(160, 210)); !changed || s.Foreground() != "chat" {
		t.Fatal("launch chat")
	}
	if _, changed := s.Input(tap(160, 236)); !changed {
		t.Fatal("send message")
	}
	if frame := s.Back(); s.Foreground() != "home" || !frame.Presents("home", 2) {
		t.Fatalf("back = foreground:%q frame:%#v", s.Foreground(), frame)
	}
}

func tap(x, y int) server.PointerEvent {
	return server.PointerEvent{
		Source:           server.TouchSource,
		Action:           server.PointerUp,
		ChangedPointerID: 0,
		Pointers:         []server.PointerSample{{ID: 0, X: x, Y: y}},
	}
}
