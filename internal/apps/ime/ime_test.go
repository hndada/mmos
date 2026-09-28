package ime

import (
	"testing"

	"mmos/internal/server"
)

func TestKeyboardComposesAndCommitsHangul(t *testing.T) {
	runtime := server.NewRuntime()
	keyboard, err := New(&runtime, server.PrimaryDisplay)
	if err != nil {
		t.Fatal(err)
	}
	frame, ok := keyboard.Show("chat")
	if !ok {
		t.Fatal("show keyboard")
	}
	if !slidesIn(frame) {
		t.Fatalf("keyboard should slide in: %#v", frame.Layers)
	}
	for _, point := range [][2]int{{112, 360}, {240, 408}} {
		event, ok := keyboard.HandleInput(tap(point[0], point[1]))
		if !ok || !event.Composing {
			t.Fatalf("compose at point=%v: %#v, handled=%t", point, event, ok)
		}
	}
	if event, ok := keyboard.HandleInput(tap(48, 408)); !ok || event.Text != "간" {
		t.Fatalf("final consonant = %#v, handled=%t", event, ok)
	}
	event, ok := keyboard.HandleInput(tap(160, 456))
	if !ok || event.Composing || event.Text != "간" {
		t.Fatalf("commit = %#v, handled=%t", event, ok)
	}
}

func slidesIn(frame server.Frame) bool {
	for _, layer := range frame.Layers {
		if layer.WindowID == window {
			return layer.OffsetY > 0
		}
	}
	return false
}

func tap(x, y int) server.PointerEvent {
	return server.PointerEvent{
		Action:           server.PointerUp,
		ChangedPointerID: 0,
		Pointers:         []server.PointerSample{{ID: 0, X: x, Y: y}},
	}
}
