package server

import (
	"testing"

	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

// FuzzRuntimeProtocol is a fault-injection harness for untrusted client
// requests. It asserts that malformed IDs, credentials, and buffer geometry
// are rejected without changing the owned window.
func FuzzRuntimeProtocol(f *testing.F) {
	f.Add("foreign", 320, 480)
	f.Add("", -1, 10)
	f.Fuzz(func(t *testing.T, id string, width, height int) {
		runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
		launch, err := runtime.Launch("chat")
		if err != nil {
			t.Fatal(err)
		}
		bounds := geom.Rect{Width: 320, Height: 480}
		if !runtime.AttachWindow(launch.Session, "owned", PrimaryDisplay, WindowConfig{Kind: AppWindow, Bounds: bounds}) ||
			!runtime.Activate(launch.Session, "owned") {
			t.Fatal("prepare owned window")
		}
		forged := Session{pid: launch.Session.PID()}
		if runtime.AttachWindow(forged, model.WindowID(id), PrimaryDisplay, WindowConfig{Kind: AppWindow, Bounds: bounds}) ||
			runtime.Submit(forged, "owned", model.Buffer{Bounds: geom.Rect{Width: width, Height: height}}) ||
			runtime.Activate(forged, "owned") || runtime.Terminate(999) {
			t.Fatal("forged session must not mutate runtime")
		}
		if runtime.Foreground(PrimaryDisplay) != "owned" {
			t.Fatalf("foreground changed to %q", runtime.Foreground(PrimaryDisplay))
		}
	})
}
