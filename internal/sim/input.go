package sim

import (
	"mmos/internal/server"
)

func (s *Simulator) Input(event server.InputEvent) (server.Frame, bool) {
	routed, ok := s.runtime.Input(displayID, event)
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	if routed.WindowID == s.homeWindow().ID {
		return s.runtime.LastFrame(displayID), s.homeApp.HandleInput(routed.WindowID, routed.Event)
	}
	if s.chat == nil || routed.WindowID != "chat" || !s.chat.HandleInput(routed.WindowID, routed.Event) {
		return s.runtime.LastFrame(displayID), false
	}
	return s.present(s.chatSession, s.chat.Process, "chat")
}
