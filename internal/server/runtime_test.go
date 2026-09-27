package server

import (
	"testing"

	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

func TestRuntimeComposesAndRoutesMultipleWindows(t *testing.T) {
	runtime := NewRuntime(
		&model.AppPackage{ID: "left", EntryPoint: "Left.Main"},
		&model.AppPackage{ID: "right", EntryPoint: "Right.Main"},
	)
	left, err := runtime.Launch("left")
	if err != nil {
		t.Fatal(err)
	}
	right, err := runtime.Launch("right")
	if err != nil {
		t.Fatal(err)
	}
	leftBounds := geom.Rect{Width: 160, Height: 480}
	rightBounds := geom.Rect{X: 160, Width: 160, Height: 480}
	if !runtime.AttachWindow(left.Session, "left", PrimaryDisplay, WindowConfig{Bounds: leftBounds}) ||
		!runtime.AttachWindow(right.Session, "right", PrimaryDisplay, WindowConfig{Bounds: rightBounds}) ||
		!runtime.Activate(left.Session, "left") || !runtime.Activate(right.Session, "right") {
		t.Fatal("prepare windows")
	}
	if !runtime.Submit(left.Session, "left", model.Buffer{Revision: 1, Bounds: leftBounds}) ||
		!runtime.Submit(right.Session, "right", model.Buffer{Revision: 1, Bounds: rightBounds}) {
		t.Fatal("submit window buffers")
	}
	frame := runtime.VSync(PrimaryDisplay)
	if len(frame.Layers) != 2 || !frame.Presents("left", 1) || !frame.Presents("right", 1) {
		t.Fatalf("frame = %#v", frame)
	}

	if _, ok := runtime.Input(PrimaryDisplay, PointerEvent{
		Action: PointerDown, ChangedPointerID: 0,
		Pointers: []PointerSample{{ID: 0, X: 20, Y: 30}},
	}); !ok {
		t.Fatal("dispatch pointer down")
	}
	routed, ok := runtime.Input(PrimaryDisplay, PointerEvent{
		Action: PointerUp, ChangedPointerID: 0,
		Pointers: []PointerSample{{ID: 0, X: 20, Y: 30}},
	})
	if !ok || routed.WindowID != "left" || runtime.Foreground(PrimaryDisplay) != "left" {
		t.Fatalf("route = %#v foreground=%q", routed, runtime.Foreground(PrimaryDisplay))
	}
	point, ok := routed.Event.(PointerEvent).PointerByID(0)
	if !ok || point.X != 20 || point.Y != 30 {
		t.Fatalf("local point = %#v found=%t", point, ok)
	}
}

func TestRuntimeRejectsWindowBoundsOutsideDisplay(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	if runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, WindowConfig{
		Bounds: geom.Rect{X: 160, Width: 161, Height: 480},
	}) {
		t.Fatal("window extending past the display must be rejected")
	}
}

func TestRuntimeCachesProcessAfterItsLastWindowIsHidden(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	bounds := geom.Rect{Width: 320, Height: 480}
	if !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, WindowConfig{Bounds: bounds}) ||
		!runtime.Activate(launch.Session, "chat") || !runtime.HideWindow(launch.Session, "chat") {
		t.Fatal("hide chat window")
	}
	if state, ok := runtime.ProcessState(launch.Session.PID()); !ok || state != Cached {
		t.Fatalf("state = %v found=%t", state, ok)
	}
}

func TestRuntimeReleasesPointerCaptureWhenWindowIsHidden(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	bounds := geom.Rect{Width: 320, Height: 480}
	if !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, WindowConfig{Bounds: bounds}) ||
		!runtime.Activate(launch.Session, "chat") {
		t.Fatal("prepare chat window")
	}
	if _, ok := runtime.Input(PrimaryDisplay, PointerEvent{
		Action: PointerDown, ChangedPointerID: 0,
		Pointers: []PointerSample{{ID: 0, X: 20, Y: 30}},
	}); !ok {
		t.Fatal("dispatch pointer down")
	}
	if !runtime.HideWindow(launch.Session, "chat") || !runtime.Activate(launch.Session, "chat") {
		t.Fatal("hide and reactivate chat window")
	}
	if _, ok := runtime.Input(PrimaryDisplay, PointerEvent{
		Action: PointerUp, ChangedPointerID: 0,
		Pointers: []PointerSample{{ID: 0, X: 20, Y: 30}},
	}); ok {
		t.Fatal("pointer capture must be released when its window is hidden")
	}
}

func TestRuntimeSplitResizesAndShowsTwoWindows(t *testing.T) {
	runtime := NewRuntime(
		&model.AppPackage{ID: "one", EntryPoint: "One.Main"},
		&model.AppPackage{ID: "two", EntryPoint: "Two.Main"},
	)
	one, _ := runtime.Launch("one")
	two, _ := runtime.Launch("two")
	full := geom.Rect{Width: 320, Height: 480}
	if !runtime.AttachWindow(one.Session, "one", PrimaryDisplay, WindowConfig{Bounds: full}) ||
		!runtime.AttachWindow(two.Session, "two", PrimaryDisplay, WindowConfig{Bounds: full}) ||
		!runtime.Split(PrimaryDisplay, "one", "two") {
		t.Fatal("split windows")
	}
	if !runtime.Submit(one.Session, "one", model.Buffer{Revision: 1, Bounds: geom.Rect{Width: 160, Height: 480}}) ||
		!runtime.Submit(two.Session, "two", model.Buffer{Revision: 1, Bounds: geom.Rect{X: 160, Width: 160, Height: 480}}) {
		t.Fatal("submit split buffers")
	}
	frame := runtime.VSync(PrimaryDisplay)
	if len(frame.Layers) != 2 || !frame.Presents("one", 1) || !frame.Presents("two", 1) {
		t.Fatalf("frame = %#v", frame)
	}
}
