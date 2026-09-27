package main

import (
	"fmt"
	"os"

	"mmos/internal/server"
	"mmos/internal/sim"
)

func main() {
	s, err := sim.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	s.Input(tap(160, 210))
	frame, changed := s.Input(tap(160, 236))
	fmt.Printf("frame=%d message_changed=%t\n", frame.Number, changed)
}

func tap(x, y int) server.PointerEvent {
	return server.PointerEvent{
		Source:           server.TouchSource,
		Action:           server.PointerUp,
		ChangedPointerID: 0,
		Pointers: []server.PointerSample{
			{ID: 0, X: x, Y: y},
		},
	}
}
