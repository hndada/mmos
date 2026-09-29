package main

import (
	"fmt"
	"os"

	"mmos/internal/server"
)

func main() {
	s, err := newSimulator()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	s.input(pointerEvent(server.PointerDown, 160, 210))
	s.input(pointerEvent(server.PointerUp, 160, 210))
	s.input(pointerEvent(server.PointerDown, 160, 236))
	frame, changed := s.input(pointerEvent(server.PointerUp, 160, 236))
	fmt.Printf("frame=%d message_changed=%t\n", frame.Number, changed)
}

func pointerEvent(action server.PointerAction, x, y int) server.PointerEvent {
	return server.PointerEvent{
		Source:           server.TouchSource,
		Action:           action,
		ChangedPointerID: 0,
		Pointers: []server.PointerSample{
			{ID: 0, X: x, Y: y},
		},
	}
}
