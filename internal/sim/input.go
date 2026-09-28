package sim

import (
	"mmos/internal/server"
)

func (s *Simulator) Input(event server.InputEvent) (server.Frame, bool) {
	if _, ok := event.(server.RotateEvent); ok {
		return s.rotate()
	}
	if system, ok := event.(server.SystemEvent); ok {
		if system.Action == server.SystemBack && s.ime.Visible() {
			_, _ = s.ime.Hide()
			if s.chat != nil {
				s.chat.State.Editing = false
				s.chat.SetKeyboardInset(0)
				return s.present(s.chatSession, s.chat.Process, "chat")
			}
			return s.runtime.LastFrame(displayID), true
		}
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
	if text, handled := s.ime.HandleInput(event); handled {
		if s.ime.Target() == "chat" && s.chat != nil && s.chat.HandleInput("chat", text) {
			return s.present(s.chatSession, s.chat.Process, "chat")
		}
		return s.runtime.LastFrame(displayID), false
	}
	routed, ok := s.runtime.Input(displayID, event)
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	if routed.WindowID == s.homeWindow().ID {
		return s.runtime.LastFrame(displayID), s.homeApp.HandleInput(routed.WindowID, routed.Event)
	}
	if s.chat != nil && routed.WindowID == "chat" {
		messages := s.chat.State.MessageCount
		if !s.chat.HandleInput(routed.WindowID, routed.Event) {
			return s.runtime.LastFrame(displayID), false
		}
		if s.chat.Editing() && !s.ime.Visible() {
			if _, ok := s.ime.Show(routed.WindowID); ok {
				s.chat.SetKeyboardInset(s.ime.Inset())
			}
		}
		if s.chat.State.MessageCount > messages {
			s.runtime.PublishNotice(s.chatSession, "Chat", "Message sent")
		}
		return s.present(s.chatSession, s.chat.Process, "chat")
	}
	if s.settings != nil && routed.WindowID == "settings" {
		if !s.settings.HandleInput(routed.WindowID, routed.Event) {
			return s.runtime.LastFrame(displayID), false
		}
		return s.present(s.settingsSession, s.settings.Process, "settings")
	}
	return s.runtime.LastFrame(displayID), false
}

func (s *Simulator) rotate() (server.Frame, bool) {
	display, ok := s.runtime.Rotate(displayID)
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	if !s.lock.Resize() || !s.notice.Resize() || !s.history.Resize() || !s.ime.Resize() {
		return s.runtime.LastFrame(displayID), false
	}
	if !s.resize(s.homeSession, s.homeApp.Process, s.homeWindow().ID, display.Bounds()) {
		return s.runtime.LastFrame(displayID), false
	}
	if s.chat != nil && !s.resize(s.chatSession, s.chat.Process, "chat", display.Bounds()) {
		return s.runtime.LastFrame(displayID), false
	}
	if s.settings != nil && !s.resize(s.settingsSession, s.settings.Process, "settings", display.Bounds()) {
		return s.runtime.LastFrame(displayID), false
	}
	if s.chat != nil && s.ime.Visible() {
		s.chat.SetKeyboardInset(s.ime.Inset())
	}
	return s.presentForeground(), true
}
