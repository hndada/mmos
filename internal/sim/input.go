package sim

import (
	"mmos/internal/server"
)

func (s *Simulator) Input(event server.InputEvent) (server.Frame, bool) {
	if _, ok := event.(server.RotateEvent); ok {
		return s.rotate()
	}
	if system, ok := event.(server.SystemEvent); ok {
		if frame, handled := s.lock.HandleSystem(system); handled {
			return frame, true
		}
		if frame, handled := s.notice.HandleSystem(system); handled {
			return frame, true
		}
		if system.Action == server.SystemRecents {
			return s.history.Show(s.historyEntries())
		}
		if system.Action == server.SystemHome {
			return s.activateHome(), true
		}
		if system.Action == server.SystemBack && s.chat != nil && s.Foreground() == "chat" && s.chat.Back() {
			return s.present(s.chatSession, s.chat.Process, "chat")
		}
		if system.Action == server.SystemBack {
			return s.activateHome(), true
		}
		return s.runtime.LastFrame(displayID), false
	}
	if frame, handled := s.lock.HandleInput(event); handled {
		return frame, true
	}
	if frame, handled := s.notice.HandleInput(event); handled {
		return frame, true
	}
	if id, frame, handled := s.history.HandleInput(event); handled {
		if id != "" {
			return s.activate(id), true
		}
		return frame, true
	}
	routed, ok := s.runtime.Input(displayID, event)
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	if routed.WindowID == s.homeWindow().ID {
		return s.runtime.LastFrame(displayID), s.homeApp.HandleInput(routed.WindowID, routed.Event)
	}
	if s.chat != nil && routed.WindowID == "chat" && s.chat.HandleInput(routed.WindowID, routed.Event) {
		s.runtime.PublishNotice(s.chatSession, "Chat", "Message sent")
		return s.present(s.chatSession, s.chat.Process, "chat")
	}
	return s.runtime.LastFrame(displayID), false
}

func (s *Simulator) rotate() (server.Frame, bool) {
	display, ok := s.runtime.Rotate(displayID)
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	if !s.lock.Resize() || !s.notice.Resize() || !s.history.Resize() {
		return s.runtime.LastFrame(displayID), false
	}
	if !s.resize(s.homeSession, s.homeApp.Process, s.homeWindow().ID, display.Bounds()) {
		return s.runtime.LastFrame(displayID), false
	}
	if s.chat != nil && !s.resize(s.chatSession, s.chat.Process, "chat", display.Bounds()) {
		return s.runtime.LastFrame(displayID), false
	}
	return s.presentForeground(), true
}
