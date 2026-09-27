package server

import (
	"testing"

	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

func TestRecordingFiltersProtectedLayers(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	bounds := geom.Rect{Width: 320, Height: 480}
	if !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, WindowConfig{Bounds: bounds}) ||
		!runtime.Activate(launch.Session, "chat") ||
		!runtime.RegisterSystemWindow("lock", PrimaryDisplay, bounds) ||
		!runtime.ProtectWindow("lock") ||
		!runtime.Submit(launch.Session, "chat", model.Buffer{Revision: 1, Bounds: bounds}) ||
		!runtime.SubmitSystem("lock", model.Buffer{Revision: 1, Bounds: bounds}) ||
		!runtime.ShowSystemWindow("lock") {
		t.Fatal("prepare recording")
	}
	if err := runtime.StartRecording(PrimaryDisplay); err != nil {
		t.Fatal(err)
	}
	runtime.VSync(PrimaryDisplay)
	recording, ok := runtime.StopRecording(PrimaryDisplay)
	if !ok || len(recording.Frames) != 1 || recording.Frames[0].Presents("lock", 1) {
		t.Fatalf("recording = %#v found=%t", recording, ok)
	}
}

func TestSystemWindowCanBeRegeneratedWithoutDisplayChange(t *testing.T) {
	runtime := NewRuntime()
	bounds := geom.Rect{Width: 320, Height: 480}
	if !runtime.RegisterSystemWindow("settings", PrimaryDisplay, bounds) ||
		!runtime.SubmitSystem("settings", model.Buffer{Revision: 1, Bounds: bounds}) ||
		!runtime.SubmitSystem("settings", model.Buffer{Revision: 1, Bounds: bounds}) ||
		!runtime.ShowSystemWindow("settings") {
		t.Fatal("regenerate system window")
	}
	frame := runtime.VSync(PrimaryDisplay)
	if !frame.Presents("settings", 2) {
		t.Fatalf("frame = %#v", frame)
	}
}

func TestAppCanProtectOnlyItsOwnContent(t *testing.T) {
	runtime := NewRuntime(
		&model.AppPackage{ID: "one", EntryPoint: "One.Main"},
		&model.AppPackage{ID: "two", EntryPoint: "Two.Main"},
	)
	one, _ := runtime.Launch("one")
	two, _ := runtime.Launch("two")
	bounds := geom.Rect{Width: 320, Height: 480}
	if !runtime.AttachWindow(one.Session, "one", PrimaryDisplay, WindowConfig{Bounds: bounds}) ||
		!runtime.AttachWindow(two.Session, "two", PrimaryDisplay, WindowConfig{Bounds: bounds}) ||
		!runtime.Activate(one.Session, "one") ||
		!runtime.Activate(two.Session, "two") ||
		!runtime.Submit(one.Session, "one", model.Buffer{Revision: 1, Bounds: bounds}) ||
		!runtime.Submit(two.Session, "two", model.Buffer{Revision: 1, Bounds: bounds}) ||
		!runtime.SetContentProtected(one.Session, "one", true) ||
		runtime.SetContentProtected(two.Session, "one", false) {
		t.Fatal("protect own content")
	}
	runtime.VSync(PrimaryDisplay)
	capture, ok := runtime.Capture(PrimaryDisplay)
	if !ok || capture.Presents("one", 1) || !capture.Presents("two", 1) {
		t.Fatalf("capture = %#v found=%t", capture, ok)
	}
}

func TestCompositorCaptureRenumbersRemainingLayers(t *testing.T) {
	compositor := Compositor{LastFrame: Frame{
		Layers: []Layer{{WindowID: "bottom", ZIndex: 0}, {WindowID: "top", ZIndex: 1}},
	}}
	capture := compositor.Capture(map[WindowID]bool{"bottom": true})
	if len(capture.Layers) != 1 || capture.Layers[0].WindowID != "top" || capture.Layers[0].ZIndex != 0 {
		t.Fatalf("capture = %#v", capture)
	}
}
