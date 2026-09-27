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
	s.Input(tap(server.PointerDown, 160, 210))
	s.Input(tap(server.PointerUp, 160, 210))
	s.Input(tap(server.PointerDown, 160, 236))
	frame, changed := s.Input(tap(server.PointerUp, 160, 236))
	fmt.Printf("frame=%d message_changed=%t\n", frame.Number, changed)
}

func tap(action server.PointerAction, x, y int) server.PointerEvent {
	return server.PointerEvent{
		Source:           server.TouchSource,
		Action:           action,
		ChangedPointerID: 0,
		Pointers: []server.PointerSample{
			{ID: 0, X: x, Y: y},
		},
	}
}
