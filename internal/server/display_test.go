package server

import (
	"testing"

	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

func TestRuntimeMovesWindowBetweenIndependentDisplays(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	external := DisplayID("external")
	if !runtime.AddDisplay(external, DisplayConfig{Width: 1920, Height: 1080, Revision: 1, Density: 1}) {
		t.Fatal("add external display")
	}
	launch, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	primary := geom.Rect{Width: 320, Height: 480}
	if !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, WindowConfig{Bounds: primary}) ||
		!runtime.Activate(launch.Session, "chat") ||
		!runtime.MoveWindow("chat", external, geom.Rect{Width: 1920, Height: 1080}) {
		t.Fatal("move chat window")
	}
	if runtime.Foreground(PrimaryDisplay) != "" || runtime.Foreground(external) != "" {
		t.Fatal("moving a window must not fabricate focus on either display")
	}
	if !runtime.Activate(launch.Session, "chat") ||
		!runtime.Submit(launch.Session, "chat", model.Buffer{Revision: 1, Bounds: geom.Rect{Width: 1920, Height: 1080}}) {
		t.Fatal("present moved window")
	}
	frame := runtime.VSync(external)
	if !frame.Presents("chat", 1) || runtime.VSync(PrimaryDisplay).Presents("chat", 1) {
		t.Fatalf("display scenes leaked: external=%#v", frame)
	}
}

func TestRuntimeRejectsStaleBuffer(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	bounds := geom.Rect{Width: 320, Height: 480}
	if !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, WindowConfig{Bounds: bounds}) ||
		!runtime.Submit(launch.Session, "chat", model.Buffer{Revision: 2, Bounds: bounds}) ||
		runtime.Submit(launch.Session, "chat", model.Buffer{Revision: 2, Bounds: bounds}) ||
		runtime.Submit(launch.Session, "chat", model.Buffer{Revision: 1, Bounds: bounds}) {
		t.Fatal("stale buffer acceptance")
	}
}

func TestRuntimeSupportsMultipleInstancesAndDividerChanges(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	first, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	second, err := runtime.LaunchInstance("chat")
	if err != nil || second.Reused || first.Session.PID() == second.Session.PID() {
		t.Fatalf("second instance = %#v, %v", second, err)
	}
	full := geom.Rect{Width: 320, Height: 480}
	if !runtime.AttachWindow(first.Session, "first", PrimaryDisplay, WindowConfig{Bounds: full}) ||
		!runtime.AttachWindow(second.Session, "second", PrimaryDisplay, WindowConfig{Bounds: full}) ||
		!runtime.Split(PrimaryDisplay, "first", "second") ||
		!runtime.ResizeSplit(PrimaryDisplay, "first", "second", 120) {
		t.Fatal("split windows")
	}
	if !runtime.Activate(first.Session, "first") || runtime.Foreground(PrimaryDisplay) != "first" {
		t.Fatal("activate first window")
	}
}

func TestDisplayRotateTransformsInsetsAndCutout(t *testing.T) {
	config := DisplayConfig{
		Width: 320, Height: 480, Revision: 1,
		SafeArea: geom.Insets{Top: 30, Right: 4, Bottom: 10, Left: 2},
		Cutout:   geom.Rect{X: 100, Y: 0, Width: 120, Height: 30},
	}
	rotated := config.Rotate()
	if rotated.Bounds() != (geom.Rect{Width: 480, Height: 320}) ||
		rotated.SafeArea != (geom.Insets{Top: 2, Right: 30, Bottom: 4, Left: 10}) ||
		rotated.Cutout != (geom.Rect{X: 290, Y: 100, Width: 30, Height: 120}) {
		t.Fatalf("rotated config = %#v", rotated)
	}
}
