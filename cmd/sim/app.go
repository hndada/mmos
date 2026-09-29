package main

import (
	"mmos/internal/apps/chat"
	"mmos/internal/apps/history"
	"mmos/internal/apps/settings"
	"mmos/internal/client"
	"mmos/internal/common/protocol"
	"mmos/internal/server"
)

// inProcessLauncher is the simulator's local transport endpoint. Replacing
// it with an IPC client leaves application code and protocol messages intact.
type inProcessLauncher struct {
	sim *simulator
}

func (l inProcessLauncher) Launch(request protocol.LaunchRequest) protocol.LaunchReply {
	return l.sim.launch(request)
}

func (s *simulator) launch(request protocol.LaunchRequest) protocol.LaunchReply {
	switch request.PackageID {
	case "chat-app":
		return s.launchChat()
	case "settings-app":
		return s.launchSettings()
	}
	return protocol.LaunchReply{}
}

func (s *simulator) launchChat() protocol.LaunchReply {
	launch, err := s.runtime.Launch("chat-app")
	if err != nil {
		return protocol.LaunchReply{}
	}
	splash, ok := s.lock.ShowSplash(launch.Package.ID)
	if !ok {
		return protocol.LaunchReply{}
	}
	s.splash = splash
	if !launch.Reused {
		s.chatSession, s.chat = launch.Session, chat.New(launch.Package)
		s.chat.Process.Config = s.runtime.Config()
		window := s.chat.Process.Windows["chat"]
		_, ok := s.runtime.Display(displayID)
		if !ok {
			return protocol.LaunchReply{}
		}
		if !s.runtime.AttachWindow(s.chatSession, window.ID, displayID, server.WindowConfig{Bounds: window.Bounds()}) {
			return protocol.LaunchReply{}
		}
	} else {
		s.chatSession = launch.Session
	}
	if s.chat == nil || !s.runtime.Activate(s.chatSession, "chat") || !s.lock.HideSplash() {
		return protocol.LaunchReply{}
	}
	_, ok = s.present(s.chatSession, s.chat.Process, "chat")
	return protocol.LaunchReply{Started: ok, Reused: ok && launch.Reused}
}

func (s *simulator) launchSettings() protocol.LaunchReply {
	launch, err := s.runtime.Launch("settings-app")
	if err != nil {
		return protocol.LaunchReply{}
	}
	splash, ok := s.lock.ShowSplash(launch.Package.ID)
	if !ok {
		return protocol.LaunchReply{}
	}
	s.splash = splash
	if !launch.Reused {
		s.settingsSession = launch.Session
		s.settings = settings.New(launch.Package, s.applySettings)
		s.settings.Configure(s.runtime.Config())
		window := s.settings.Process.Windows["settings"]
		if !s.runtime.AttachWindow(s.settingsSession, window.ID, displayID, server.WindowConfig{Bounds: window.Bounds()}) {
			return protocol.LaunchReply{}
		}
	} else {
		s.settingsSession = launch.Session
	}
	if s.settings == nil || !s.runtime.Activate(s.settingsSession, "settings") || !s.lock.HideSplash() {
		return protocol.LaunchReply{}
	}
	_, ok = s.present(s.settingsSession, s.settings.Process, "settings")
	return protocol.LaunchReply{Started: ok, Reused: ok && launch.Reused}
}

func (s *simulator) applySettings(config server.SystemConfig) bool {
	s.runtime.SetConfig(config)
	s.applyConfig(s.runtime.Config())
	return true
}

func (s *simulator) activateHome() server.Frame {
	if !s.runtime.Activate(s.homeSession, s.homeWindow().ID) {
		return s.runtime.LastFrame(displayID)
	}
	frame, ok := s.present(s.homeSession, s.homeApp.Process, s.homeWindow().ID)
	if !ok {
		return s.runtime.LastFrame(displayID)
	}
	return frame
}

func (s *simulator) activate(id server.WindowID) server.Frame {
	if id == s.homeWindow().ID {
		return s.activateHome()
	}
	if id == "chat" && s.chat != nil && s.runtime.Activate(s.chatSession, id) {
		frame, _ := s.present(s.chatSession, s.chat.Process, id)
		return frame
	}
	if id == "settings" && s.settings != nil && s.runtime.Activate(s.settingsSession, id) {
		frame, _ := s.present(s.settingsSession, s.settings.Process, id)
		return frame
	}
	return s.runtime.LastFrame(displayID)
}

func (s *simulator) historyEntries() []history.Entry {
	entries := []history.Entry{{WindowID: s.homeWindow().ID, Label: "Home"}}
	if s.chat != nil {
		entries = append([]history.Entry{{WindowID: "chat", Label: "Chat"}}, entries...)
	}
	if s.settings != nil {
		entries = append([]history.Entry{{WindowID: "settings", Label: "Settings"}}, entries...)
	}
	return entries
}

// back restores the most recently active app window through the server.
func (s *simulator) back() server.Frame {
	frame, _ := s.input(server.SystemEvent{Action: server.SystemBack})
	return frame
}
func (s *simulator) present(session server.Session, process *client.AppProcess, id server.WindowID) (server.Frame, bool) {
	window, ok := process.Windows[id]
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	s.drawer.Draw(window)
	return s.runtime.Present(session, id, window.Buffer)
}
func (s *simulator) resize(session server.Session, process *client.AppProcess, id server.WindowID, bounds server.Rect) bool {
	window, ok := process.Windows[id]
	if !ok || !s.runtime.ResizeWindow(session, id, bounds) {
		return false
	}
	window.Resize(bounds)
	return true
}
func (s *simulator) presentForeground() server.Frame {
	if s.runtime.Foreground(displayID) == "chat" && s.chat != nil {
		frame, _ := s.present(s.chatSession, s.chat.Process, "chat")
		return frame
	}
	if s.runtime.Foreground(displayID) == "settings" && s.settings != nil {
		frame, _ := s.present(s.settingsSession, s.settings.Process, "settings")
		return frame
	}
	frame, _ := s.present(s.homeSession, s.homeApp.Process, s.homeWindow().ID)
	return frame
}
func (s *simulator) homeWindow() *client.Window { return s.homeApp.Process.Windows["home"] }
func (s *simulator) applyConfig(config server.SystemConfig) {
	s.homeApp.Process.Config = config
	if s.chat != nil {
		s.chat.Process.Config = config
	}
	if s.settings != nil {
		s.settings.Configure(config)
	}
}
