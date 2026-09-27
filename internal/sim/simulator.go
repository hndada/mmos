package sim

import (
	"fmt"

	"mmos/internal/app/chat"
	"mmos/internal/app/home"
	"mmos/internal/client"
	"mmos/internal/server"
)

const (
	displayID = server.PrimaryDisplay
)

// Simulator wires one display, trusted system UI, and installed applications.
type Simulator struct {
	runtime     server.Runtime
	drawer      client.Drawer
	homeApp     *home.App
	homeSession server.Session
	chat        *chat.App
	chatSession server.Session
}

func New() (*Simulator, error) {
	runtime := server.NewRuntime(
		&server.AppPackage{ID: "home", EntryPoint: "Home.Main"},
		&server.AppPackage{ID: "chat-app", EntryPoint: "Chat.Main"},
	)
	homeLaunch, err := runtime.Launch("home")
	if err != nil {
		return nil, fmt.Errorf("launch home: %w", err)
	}
	s := &Simulator{runtime: runtime, homeSession: homeLaunch.Session}
	s.homeApp = home.New(homeLaunch.Package, s.LaunchChat)
	s.applyConfig(runtime.Config())
	window := s.homeWindow()
	if !s.runtime.AttachWindow(s.homeSession, window.ID, displayID, server.WindowConfig{Bounds: window.Bounds()}) || !s.runtime.Activate(s.homeSession, window.ID) || !s.draw(s.homeSession, s.homeApp.Process, window.ID) {
		return nil, fmt.Errorf("start home window")
	}
	return s, nil
}

func (s *Simulator) Terminate(pid int) server.Frame {
	if pid == s.homeSession.PID() || !s.runtime.Terminate(pid) {
		return s.runtime.LastFrame(displayID)
	}
	if s.chat != nil && s.chatSession.PID() == pid {
		s.chat, s.chatSession = nil, server.Session{}
	}
	return s.activateHome()
}
func (s *Simulator) Frame() server.Frame         { return s.runtime.LastFrame(displayID) }
func (s *Simulator) Foreground() server.WindowID { return s.runtime.Foreground(displayID) }
