package server

import (
	"mmos/internal/common/geom"
)

// WindowConfig describes the server-visible properties of a client window.
type WindowConfig struct {
	Bounds geom.Rect
}

// Window is the server-side record for one client window. It deliberately has
// no UI tree: the server receives only a submitted pixel Buffer.
type Window struct {
	ID      WindowID
	Bounds  geom.Rect
	Visible bool
	Buffer  Buffer
}

// WindowServer owns registration, visibility, focus, and z-order.
type WindowServer struct {
	Windows    map[WindowID]*Window
	Foreground WindowID
	ZOrder     []WindowID
	history    []WindowID
}

func NewWindowServer() WindowServer {
	return WindowServer{Windows: map[WindowID]*Window{}}
}

func (s *WindowServer) Register(id WindowID, config WindowConfig) {
	s.Windows[id] = &Window{ID: id, Bounds: config.Bounds}
	s.ZOrder = append(s.ZOrder, id)
}

func (s *WindowServer) Unregister(id WindowID) {
	delete(s.Windows, id)
	for index, candidate := range s.ZOrder {
		if candidate == id {
			s.ZOrder = append(s.ZOrder[:index], s.ZOrder[index+1:]...)
			break
		}
	}
	if s.Foreground == id {
		s.Foreground = ""
	}
	for i := len(s.history) - 1; i >= 0; i-- {
		if s.history[i] == id {
			s.history = append(s.history[:i], s.history[i+1:]...)
		}
	}
}

func (s *WindowServer) Activate(id WindowID) bool {
	window, ok := s.Windows[id]
	if !ok {
		return false
	}
	if s.Foreground != "" && s.Foreground != id {
		s.history = append(s.history, s.Foreground)
	}
	s.show(id)
	s.Foreground = window.ID
	return true
}

// Back restores the most recently foregrounded remaining window.
func (s *WindowServer) Back() (WindowID, bool) {
	for len(s.history) > 0 {
		index := len(s.history) - 1
		id := s.history[index]
		s.history = s.history[:index]
		window, ok := s.Windows[id]
		if !ok {
			continue
		}
		s.show(id)
		s.Foreground = window.ID
		return id, true
	}
	return "", false
}

func (s *WindowServer) show(id WindowID) {
	s.moveAppToFront(id)
}

func (s *WindowServer) moveAppToFront(id WindowID) {
	for _, window := range s.Windows {
		window.Visible = window.ID == id
	}
	s.removeFromZOrder(id)
	s.ZOrder = append(s.ZOrder, id)
}

func (s *WindowServer) removeFromZOrder(id WindowID) {
	for index, candidate := range s.ZOrder {
		if candidate == id {
			s.ZOrder = append(s.ZOrder[:index], s.ZOrder[index+1:]...)
			return
		}
	}
}

// SubmitBuffer copies a client-produced Buffer into the server record.
func (s *WindowServer) SubmitBuffer(id WindowID, buffer Buffer) bool {
	window, ok := s.Windows[id]
	if !ok {
		return false
	}
	window.Buffer = buffer
	return true
}

func (s WindowServer) FocusedWindowID() (WindowID, bool) {
	if s.Foreground == "" {
		return "", false
	}
	window, ok := s.Windows[s.Foreground]
	if !ok || !window.Visible {
		return "", false
	}
	return window.ID, true
}
