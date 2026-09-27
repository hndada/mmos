package sim

import (
	"fmt"
	"mmos/internal/server"
)

func (s *Simulator) systemEvent(event server.SystemEvent) (server.Frame, bool) {
	if s.locked && event.Action != server.SystemUnlock && event.Action != server.SystemScreenOn && event.Action != server.SystemScreenOff {
		return s.runtime.LastFrame(displayID), false
	}
	switch event.Action {
	case server.SystemBack:
		if s.settings {
			return s.hideSettings(), true
		}
		return s.activateHome(), true
	case server.SystemHome:
		return s.activateHome(), true
	case server.SystemLock:
		return s.showLock(), true
	case server.SystemUnlock:
		return s.unlock(), true
	case server.SystemSettings:
		if s.settings {
			return s.hideSettings(), true
		}
		return s.showSettings(), true
	case server.SystemScreenOff:
		return s.screenOff(), true
	case server.SystemScreenOn:
		return s.screenOn(), true
	default:
		return s.runtime.LastFrame(displayID), false
	}
}
func (s *Simulator) showLock() server.Frame {
	if !s.showOverlay(lockWindow, "lock") {
		return s.runtime.LastFrame(displayID)
	}
	s.locked = true
	return s.runtime.VSync(displayID)
}
func (s *Simulator) unlock() server.Frame {
	display, ok := s.runtime.Display(displayID)
	if !ok || display.Power == server.ScreenOff || !s.runtime.HideSystemWindow(lockWindow) {
		return s.runtime.LastFrame(displayID)
	}
	s.locked = false
	return s.runtime.VSync(displayID)
}
func (s *Simulator) screenOff() server.Frame {
	s.showLock()
	if _, ok := s.runtime.SetScreenPower(displayID, server.ScreenOff); !ok {
		return s.runtime.LastFrame(displayID)
	}
	return s.runtime.VSync(displayID)
}
func (s *Simulator) screenOn() server.Frame {
	if _, ok := s.runtime.SetScreenPower(displayID, server.ScreenOn); !ok {
		return s.runtime.LastFrame(displayID)
	}
	return s.runtime.VSync(displayID)
}
func (s *Simulator) showSettings() server.Frame {
	if !s.showOverlay(settingsWindow, s.settingsContent()) {
		return s.runtime.LastFrame(displayID)
	}
	s.settings = true
	return s.runtime.VSync(displayID)
}
func (s *Simulator) hideSettings() server.Frame {
	if !s.runtime.HideSystemWindow(settingsWindow) {
		return s.runtime.LastFrame(displayID)
	}
	s.settings = false
	return s.runtime.VSync(displayID)
}
func (s *Simulator) settingsInput(event server.InputEvent) (server.Frame, bool) {
	pointer, ok := event.(server.PointerEvent)
	if !ok || pointer.Action != server.PointerUp || pointer.ChangedPointerID != 0 {
		return s.runtime.LastFrame(displayID), false
	}
	config := s.runtime.Config()
	if config.Theme == server.LightTheme {
		config.Theme = server.DarkTheme
	} else {
		config.Theme = server.LightTheme
	}
	s.applyConfig(s.runtime.SetConfig(config))
	if !s.showOverlay(settingsWindow, s.settingsContent()) {
		return s.runtime.LastFrame(displayID), false
	}
	return s.presentForeground(), true
}
func (s *Simulator) settingsContent() string {
	return fmt.Sprintf("settings(theme=%d)", s.runtime.Config().Theme)
}
