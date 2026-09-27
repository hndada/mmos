package server

import (
	"bytes"
	"testing"

	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

func TestRuntimeRestoresSavedState(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	first, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	if !runtime.SaveState(first.Session, []byte("draft")) {
		t.Fatal("save state")
	}
	bounds := geom.Rect{Width: 320, Height: 480}
	if !runtime.AttachWindow(first.Session, "chat", PrimaryDisplay, WindowConfig{Bounds: bounds}) || !runtime.Activate(first.Session, "chat") || !runtime.HideWindow(first.Session, "chat") || !runtime.Evict(first.Session.PID()) {
		t.Fatal("evict process")
	}
	restarted, err := runtime.Launch("chat")
	if err != nil || restarted.Reused || !bytes.Equal(restarted.State, []byte("draft")) {
		t.Fatalf("restart = %#v, %v", restarted, err)
	}
}
