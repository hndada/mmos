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

func tap(x, y int) model.PointerEvent {
	return model.PointerEvent{Action: model.PointerUp, ChangedPointerID: 0, Pointers: []model.PointerSample{{ID: 0, X: x, Y: y}}}
}
