package server

import (
	"testing"
	"time"
)

func TestSlideSystemWindowOffsetsFirstFrame(t *testing.T) {
	runtime := NewRuntime()
	bounds := runtime.displays[PrimaryDisplay].config.Bounds()
	if !runtime.RegisterSystemWindow("ime", PrimaryDisplay, bounds) ||
		!runtime.SubmitSystem("ime", Buffer{Revision: 1, Bounds: bounds}) ||
		!runtime.ShowSystemWindow("ime") ||
		!runtime.SlideSystemWindow("ime", 144, time.Second) {
		t.Fatal("start system slide")
	}
	frame := runtime.VSync(PrimaryDisplay)
	for _, layer := range frame.Layers {
		if layer.WindowID == "ime" && layer.OffsetY > 0 {
			return
		}
	}
	t.Fatalf("ime slide = %#v", frame.Layers)
}

func TestSlideSystemWindowHonorsReducedMotion(t *testing.T) {
	runtime := NewRuntime()
	config := runtime.Config()
	config.ReducedMotion = true
	runtime.SetConfig(config)
	bounds := runtime.displays[PrimaryDisplay].config.Bounds()
	if !runtime.RegisterSystemWindow("ime", PrimaryDisplay, bounds) ||
		!runtime.SubmitSystem("ime", Buffer{Revision: 1, Bounds: bounds}) ||
		!runtime.ShowSystemWindow("ime") ||
		!runtime.SlideSystemWindow("ime", 144, time.Second) {
		t.Fatal("start reduced-motion system slide")
	}
	for _, layer := range runtime.VSync(PrimaryDisplay).Layers {
		if layer.WindowID == "ime" && layer.OffsetY != 0 {
			t.Fatalf("reduced motion offset = %d", layer.OffsetY)
		}
	}
}
