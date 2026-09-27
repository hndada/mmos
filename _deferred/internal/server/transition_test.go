package server

import (
	"testing"
	"time"

	"mmos/internal/common/model"
)

func TestTransitionFadesPreviousBufferIntoTarget(t *testing.T) {
	timeNow := time.Now()
	transition := transition{
		from: Layer{
			WindowID: "home",
			Buffer:   model.Buffer{Revision: 4, Content: "home"},
		},
		to:      "chat",
		started: timeNow.Add(-transitionDuration / 2),
	}
	layers := transition.apply([]Layer{{
		WindowID: "chat",
		Buffer:   model.Buffer{Revision: 7, Content: "chat"},
		Opacity:  1,
	}}, timeNow)
	if len(layers) != 2 {
		t.Fatalf("layers = %#v, want previous and target", layers)
	}
	if layers[0].WindowID != "home" || layers[0].Buffer.Revision != 4 ||
		layers[0].Opacity <= 0 || layers[0].Opacity >= 1 {
		t.Fatalf("previous layer = %#v", layers[0])
	}
	if layers[1].WindowID != "chat" || layers[1].Buffer.Revision != 7 ||
		layers[1].Opacity <= 0 || layers[1].Opacity >= 1 {
		t.Fatalf("target layer = %#v", layers[1])
	}
	if layers[0].ZIndex != 0 || layers[1].ZIndex != 1 {
		t.Fatalf("z-order = %#v", layers)
	}
}

func TestTransitionExpiresAfterDuration(t *testing.T) {
	transition := transition{
		from:    Layer{WindowID: "home"},
		to:      "chat",
		started: time.Now().Add(-transitionDuration),
	}
	layers := transition.apply(
		[]Layer{{WindowID: "chat", Opacity: 1}},
		time.Now(),
	)
	if transition.active() {
		t.Fatal("completed transition must clear its state")
	}
	if len(layers) != 1 || layers[0].WindowID != "chat" || layers[0].Opacity != 1 {
		t.Fatalf("completed layers = %#v", layers)
	}
}
