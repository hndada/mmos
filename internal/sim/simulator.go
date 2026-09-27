package sim

import (
	"fmt"

	"mmos/internal/apps/chat"
	"mmos/internal/apps/history"
	"mmos/internal/apps/home"
	"mmos/internal/apps/lock"
	"mmos/internal/apps/notice"
	"mmos/internal/client"
	"mmos/internal/server"
)

const displayID = server.PrimaryDisplay

// Simulator wires one display, trusted system UI, and installed applications.
type Simulator struct {
	runtime     server.Runtime
	drawer      client.Drawer
	lock        *lock.App
	notice      *notice.App
	history     *history.App
	homeApp     *home.App
	homeSession server.Session
	chat        *chat.App
	chatSession server.Session
	splash      server.Frame
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
	locks, err := lock.New(&s.runtime, displayID)
	if err != nil {
		return nil, fmt.Errorf("install lock: %w", err)
	}
	notices, err := notice.New(&s.runtime, displayID)
	if err != nil {
		return nil, fmt.Errorf("install notice: %w", err)
	}
	recent, err := history.New(&s.runtime, displayID)
	if err != nil {
		return nil, fmt.Errorf("install history: %w", err)
	}
	s.lock, s.notice, s.history = locks, notices, recent
	splash, ok := s.lock.ShowSplash(homeLaunch.Package.ID)
	if !ok {
		return nil, fmt.Errorf("show home splash")
	}
	s.splash = splash
	window := s.homeWindow()
	if !s.runtime.AttachWindow(s.homeSession, window.ID, displayID, server.WindowConfig{Bounds: window.Bounds()}) || !s.runtime.Activate(s.homeSession, window.ID) || !s.lock.HideSplash() {
		return nil, fmt.Errorf("start home window")
	}
	if _, ok := s.present(s.homeSession, s.homeApp.Process, window.ID); !ok {
		return nil, fmt.Errorf("present home window")
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
func (s *Simulator) SplashFrame() server.Frame   { return s.splash }
func (s *Simulator) Foreground() server.WindowID { return s.runtime.Foreground(displayID) }
