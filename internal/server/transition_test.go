package server

import (
	"testing"
	"time"
)

func TestTransitionFadesPreviousBufferIntoTarget(t *testing.T) {
	now := time.Now()
	transition := transition{
		from: Layer{WindowID: "home", Buffer: Buffer{Revision: 4, Content: "home"}},
		to:   "chat", started: now.Add(-transitionDuration / 2),
	}
	layers := transition.apply([]Layer{{WindowID: "chat", Buffer: Buffer{Revision: 7, Content: "chat"}, Opacity: 1}}, now)
	if len(layers) != 2 || layers[0].WindowID != "home" || layers[1].WindowID != "chat" ||
		layers[0].Opacity <= 0 || layers[0].Opacity >= 1 || layers[1].Opacity <= 0 || layers[1].Opacity >= 1 {
		t.Fatalf("layers = %#v", layers)
	}
}
