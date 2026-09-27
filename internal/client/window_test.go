package client_test

import (
	"testing"

	"mmos/internal/app/chat"
	"mmos/internal/common/model"
)

func TestWindowRoutesOnlyVisibleEnabledCommands(t *testing.T) {
	app := chat.New(&model.AppPackage{ID: "chat-app"})
	window := app.Process.Windows["chat"]
	event := model.PointerEvent{
		Action:           model.PointerUp,
		Source:           model.TouchSource,
		ChangedPointerID: 0,
		Pointers:         []model.PointerSample{{ID: 0, X: 160, Y: 236}},
	}

	if command, ok := window.CommandAt(event); !ok || command != chat.SendMessage {
		t.Fatal("visible, enabled Send button must receive its tap")
	}

	button := window.Node("send")
	button.Enabled = false
	if _, ok := window.CommandAt(event); ok {
		t.Fatal("disabled button must not emit a command")
	}

	button.Enabled = true
	button.Visible = false
	if _, ok := window.CommandAt(event); ok {
		t.Fatal("hidden button must not be a hit-test target")
	}
}

func TestWindowDoesNotTreatOtherMobileInputAsButtonActivation(t *testing.T) {
	app := chat.New(&model.AppPackage{ID: "chat-app"})
	window := app.Process.Windows["chat"]

	for _, event := range []model.InputEvent{
		model.PointerEvent{
			Action:           model.PointerDown,
			Source:           model.TouchSource,
			ChangedPointerID: 0,
			Pointers:         []model.PointerSample{{ID: 0, X: 160, Y: 236}},
		},
		model.PointerEvent{
			Action:           model.PointerMove,
			Source:           model.StylusSource,
			ChangedPointerID: 0,
			Pointers:         []model.PointerSample{{ID: 0, X: 160, Y: 236}},
		},
		model.KeyEvent{Action: model.KeyDown, Key: "Enter"},
		model.TextEvent{Text: "hello"},
		model.ScrollEvent{Source: model.TouchSource, X: 160, Y: 236, DeltaY: 40},
	} {
		if _, ok := window.CommandAt(event); ok {
			t.Fatalf("%T must not activate a button", event)
		}
	}
}
