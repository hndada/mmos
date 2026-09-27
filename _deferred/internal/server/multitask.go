package server

import (
	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

// Split is the former 2권 placement rule. It is archived, not built.
func (s *WindowServer) Split(first, second model.WindowID, firstBounds, secondBounds geom.Rect) bool {
	a, ok := s.Windows[first]
	if !ok || a.Kind != AppWindow {
		return false
	}
	b, ok := s.Windows[second]
	if !ok || b.Kind != AppWindow || first == second {
		return false
	}
	for _, window := range s.Windows {
		if window.Kind == AppWindow {
			window.Visible = window.ID == first || window.ID == second
		}
	}
	a.Bounds = firstBounds
	b.Bounds = secondBounds
	s.moveAppToFront(first)
	s.moveAppToFront(second)
	s.Foreground = first
	return true
}

// FocusAt is the former split-pane focus policy.
func (s *WindowServer) FocusAt(x, y int) bool {
	for i := len(s.ZOrder) - 1; i >= 0; i-- {
		window := s.Windows[s.ZOrder[i]]
		if window.Kind == AppWindow && window.Visible && window.Bounds.Contains(x, y) {
			s.moveAppToFront(window.ID)
			s.Foreground = window.ID
			return true
		}
	}
	return false
}
