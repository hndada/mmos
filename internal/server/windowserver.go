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

// WindowServer owns registration, visibility, focus, and z-order. It permits
// several visible windows; Foreground identifies only the input target.
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
	s.Detach(id)
}

// Detach removes a window from this scene and returns its server-owned record.
// Runtime uses it to move a window to another display without changing its
// app ownership.
func (s *WindowServer) Detach(id WindowID) *Window {
	window := s.Windows[id]
	if window == nil {
		return nil
	}
	delete(s.Windows, id)
	for index, candidate := range s.ZOrder {
		if candidate == id {
			s.ZOrder = append(s.ZOrder[:index], s.ZOrder[index+1:]...)
			break
		}
	}
	if s.Foreground == id {
		s.Foreground = s.topVisible()
	}
	for i := len(s.history) - 1; i >= 0; i-- {
		if s.history[i] == id {
			s.history = append(s.history[:i], s.history[i+1:]...)
		}
	}
	return window
}

func (s *WindowServer) Attach(window *Window) {
	s.Windows[window.ID] = window
	s.ZOrder = append(s.ZOrder, window.ID)
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
	if window, ok := s.Windows[id]; ok {
		window.Visible = true
	}
	s.moveAppToFront(id)
}

func (s *WindowServer) moveAppToFront(id WindowID) {
	s.removeFromZOrder(id)
	s.ZOrder = append(s.ZOrder, id)
}

// Hide removes a window from presentation and moves focus to the next visible
// window. Its buffer is retained so its owner can show it again later.
func (s *WindowServer) Hide(id WindowID) bool {
	window, ok := s.Windows[id]
	if !ok {
		return false
	}
	window.Visible = false
	if s.Foreground == id {
		s.Foreground = s.topVisible()
	}
	return true
}

// Forget removes an ID from focus history. Trusted overlays use it so a
// dismissed system surface cannot become an application's Back destination.
func (s *WindowServer) Forget(id WindowID) {
	for i := len(s.history) - 1; i >= 0; i-- {
		if s.history[i] == id {
			s.history = append(s.history[:i], s.history[i+1:]...)
		}
	}
}

// WindowAt returns the topmost visible window containing the display point.
// It does not change focus or z-order.
func (s WindowServer) WindowAt(x, y int) (WindowID, bool) {
	for i := len(s.ZOrder) - 1; i >= 0; i-- {
		id := s.ZOrder[i]
		window := s.Windows[id]
		if window != nil && window.Visible && window.Bounds.Contains(x, y) {
			return id, true
		}
	}
	return "", false
}

// FocusAt promotes the topmost visible window containing the display point.
func (s *WindowServer) FocusAt(x, y int) (WindowID, bool) {
	id, ok := s.WindowAt(x, y)
	if !ok {
		return "", false
	}
	if id != s.Foreground {
		if s.Foreground != "" {
			s.history = append(s.history, s.Foreground)
		}
		s.moveAppToFront(id)
		s.Foreground = id
	}
	return id, true
}

// Resize changes the server-visible bounds of a window. Its old buffer is no
// longer valid because the owner must draw for the new extent.
func (s *WindowServer) Resize(id WindowID, bounds geom.Rect) bool {
	window, ok := s.Windows[id]
	if !ok {
		return false
	}
	window.Bounds = bounds
	window.Buffer = Buffer{}
	return true
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
	if !ok || buffer.Revision <= window.Buffer.Revision {
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

func (s WindowServer) topVisible() WindowID {
	for i := len(s.ZOrder) - 1; i >= 0; i-- {
		window := s.Windows[s.ZOrder[i]]
		if window != nil && window.Visible {
			return window.ID
		}
	}
	return ""
}
