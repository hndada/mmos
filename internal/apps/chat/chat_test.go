package chat

import (
	"testing"

	"mmos/internal/common/model"
)

func TestNavigationStackHandlesBackBeforeSystemPolicy(t *testing.T) {
	app := New(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	if !app.HandleInput("chat", tap(160, 304)) || app.Page() != "details" {
		t.Fatalf("open details page = %q", app.Page())
	}
	if !app.Back() || app.Page() != "chat" || app.Back() {
		t.Fatalf("back stack page = %q", app.Page())
	}
}

func TestTextInputKeepsPreeditOutOfCommittedDraft(t *testing.T) {
	app := New(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	if !app.HandleInput("chat", tap(40, 150)) {
		t.Fatal("focus editor")
	}
	if !app.HandleInput("chat", model.TextEvent{Text: "가", Composing: true}) || app.State.Draft != "" || app.State.Composition != "가" {
		t.Fatalf("preedit = %#v", app.State)
	}
	if !app.HandleInput("chat", model.TextEvent{Text: "가"}) || app.State.Draft != "가" || app.State.Composition != "" {
		t.Fatalf("commit = %#v", app.State)
	}
}

func TestKeyboardInsetMovesComposerAboveKeyboard(t *testing.T) {
	app := New(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	app.SetKeyboardInset(144)
	window := app.Process.Windows["chat"]
	if editor, send := window.Node("editor"), window.Node("send"); editor.Bounds.Y+editor.Bounds.Height > 336 || send.Bounds.Y+send.Bounds.Height > 336 {
		t.Fatalf("composer overlaps keyboard: editor=%#v send=%#v", editor.Bounds, send.Bounds)
	}
	app.SetKeyboardInset(0)
	if editor, send := window.Node("editor"), window.Node("send"); editor.Bounds.Y != 130 || send.Bounds.Y != 200 {
		t.Fatalf("composer was not restored: editor=%#v send=%#v", editor.Bounds, send.Bounds)
	}
}

func tap(x, y int) model.PointerEvent {
	return model.PointerEvent{Action: model.PointerUp, ChangedPointerID: 0, Pointers: []model.PointerSample{{ID: 0, X: x, Y: y}}}
}
