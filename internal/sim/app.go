package sim

import (
	"mmos/internal/apps/chat"
	"mmos/internal/apps/history"
	"mmos/internal/client"
	"mmos/internal/server"
)

func (s *Simulator) LaunchChat() bool {
	launch, err := s.runtime.Launch("chat-app")
	if err != nil {
		return false
	}
	splash, ok := s.lock.ShowSplash(launch.Package.ID)
	if !ok {
		return false
	}
	s.splash = splash
	if !launch.Reused {
		s.chatSession, s.chat = launch.Session, chat.New(launch.Package)
		s.chat.Process.Configure(s.runtime.Config())
		window := s.chat.Process.Windows["chat"]
		_, ok := s.runtime.Display(displayID)
		if !ok {
			return false
		}
		if !s.runtime.AttachWindow(s.chatSession, window.ID, displayID, server.WindowConfig{Bounds: window.Bounds()}) {
			return false
		}
	} else {
		s.chatSession = launch.Session
	}
	if s.chat == nil || !s.runtime.Activate(s.chatSession, "chat") || !s.lock.HideSplash() {
		return false
	}
	_, ok = s.present(s.chatSession, s.chat.Process, "chat")
	return ok
}

func (s *Simulator) activateHome() server.Frame {
	if !s.runtime.Activate(s.homeSession, s.homeWindow().ID) {
		return s.runtime.LastFrame(displayID)
	}
	frame, ok := s.present(s.homeSession, s.homeApp.Process, s.homeWindow().ID)
	if !ok {
		return s.runtime.LastFrame(displayID)
	}
	return frame
}

func (s *Simulator) activate(id server.WindowID) server.Frame {
	if id == s.homeWindow().ID {
		return s.activateHome()
	}
	if id == "chat" && s.chat != nil && s.runtime.Activate(s.chatSession, id) {
		frame, _ := s.present(s.chatSession, s.chat.Process, id)
		return frame
	}
	return s.runtime.LastFrame(displayID)
}

func (s *Simulator) historyEntries() []history.Entry {
	entries := []history.Entry{{WindowID: s.homeWindow().ID, Label: "Home"}}
	if s.chat != nil {
		entries = append([]history.Entry{{WindowID: "chat", Label: "Chat"}}, entries...)
	}
	return entries
}

// Back restores the most recently active app window through the server.
func (s *Simulator) Back() server.Frame {
	frame, _ := s.Input(server.SystemEvent{Action: server.SystemBack})
	return frame
}
func (s *Simulator) draw(session server.Session, process *client.AppProcess, id server.WindowID) bool {
	window, ok := process.Windows[id]
	if !ok {
		return false
	}
	s.drawer.Draw(window)
	return s.runtime.Submit(session, id, window.Buffer)
}
func (s *Simulator) present(session server.Session, process *client.AppProcess, id server.WindowID) (server.Frame, bool) {
	window, ok := process.Windows[id]
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	s.drawer.Draw(window)
	return s.runtime.Present(session, id, window.Buffer)
}
func (s *Simulator) resize(session server.Session, process *client.AppProcess, id server.WindowID, bounds server.Rect) bool {
	window, ok := process.Windows[id]
	if !ok || !s.runtime.ResizeWindow(session, id, bounds) {
		return false
	}
	window.Resize(bounds)
	return true
}
func (s *Simulator) presentForeground() server.Frame {
	if s.runtime.Foreground(displayID) == "chat" && s.chat != nil {
		frame, _ := s.present(s.chatSession, s.chat.Process, "chat")
		return frame
	}
	frame, _ := s.present(s.homeSession, s.homeApp.Process, s.homeWindow().ID)
	return frame
}
func (s *Simulator) homeWindow() *client.Window { return s.homeApp.Process.Windows["home"] }
func (s *Simulator) applyConfig(config server.SystemConfig) {
	s.homeApp.Process.Configure(config)
	if s.chat != nil {
		s.chat.Process.Configure(config)
	}
}
