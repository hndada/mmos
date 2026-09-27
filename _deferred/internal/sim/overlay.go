package sim

import (
	"fmt"
	"mmos/internal/common/geom"
	"mmos/internal/server"
)

func (s *Simulator) registerSystemWindows() bool {
	status := geom.Rect{Width: 320, Height: 24}
	if !s.runtime.RegisterSystemWindow(statusWindow, displayID, status) || !s.runtime.SubmitSystem(statusWindow, server.Buffer{Revision: 1, Content: "status", Bounds: status}) {
		return false
	}
	return s.registerOverlay(lockWindow) && s.registerOverlay(settingsWindow)
}
func (s *Simulator) registerSplash() bool {
	display, ok := s.runtime.Display(displayID)
	return ok && s.runtime.RegisterSystemWindow(splashWindow, displayID, display.Bounds()) && s.runtime.HideSystemWindow(splashWindow)
}
func (s *Simulator) registerOverlay(id server.WindowID) bool {
	display, ok := s.runtime.Display(displayID)
	return ok && s.runtime.RegisterSystemWindow(id, displayID, display.Bounds()) && s.runtime.HideSystemWindow(id)
}
func (s *Simulator) showOverlay(id server.WindowID, content string) bool {
	display, ok := s.runtime.Display(displayID)
	return ok && s.runtime.SubmitSystem(id, server.Buffer{Revision: display.Revision, Content: content, Bounds: display.Bounds()}) && s.runtime.ShowSystemWindow(id)
}
func (s *Simulator) showSplash(packageID string) bool {
	display, ok := s.runtime.Display(displayID)
	if !ok {
		return false
	}
	buffer := server.Buffer{Revision: display.Revision, Content: fmt.Sprintf("splash(package=%s)", packageID), Bounds: display.Bounds()}
	if !s.runtime.SubmitSystem(splashWindow, buffer) || !s.runtime.ShowSystemWindow(splashWindow) {
		return false
	}
	s.splash = s.runtime.VSync(displayID)
	return s.splash.Presents(splashWindow, buffer.Revision)
}
