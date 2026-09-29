package main

import (
	"fmt"

	"mmos/internal/apps/chat"
	"mmos/internal/apps/history"
	"mmos/internal/apps/home"
	"mmos/internal/apps/ime"
	"mmos/internal/apps/lock"
	"mmos/internal/apps/notice"
	"mmos/internal/apps/settings"
	"mmos/internal/client"
	"mmos/internal/server"
)

const displayID = server.PrimaryDisplay

// simulator wires one display, trusted system UI, and installed applications.
type simulator struct {
	runtime         server.Runtime
	drawer          client.Drawer
	lock            *lock.App
	notice          *notice.App
	history         *history.App
	ime             *ime.App
	homeApp         *home.App
	homeSession     server.Session
	chat            *chat.App
	chatSession     server.Session
	settings        *settings.App
	settingsSession server.Session
	splash          server.Frame
}

func newSimulator() (*simulator, error) {
	runtime := server.NewRuntime(
		&server.AppPackage{ID: "home", EntryPoint: "Home.Main"},
		&server.AppPackage{ID: "chat-app", EntryPoint: "Chat.Main"},
		&server.AppPackage{ID: "settings-app", EntryPoint: "Settings.Main"},
	)
	homeLaunch, err := runtime.Launch("home")
	if err != nil {
		return nil, fmt.Errorf("launch home: %w", err)
	}
	s := &simulator{runtime: runtime, homeSession: homeLaunch.Session}
	s.homeApp = home.New(homeLaunch.Package, inProcessLauncher{sim: s})
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
	keyboard, err := ime.New(&s.runtime, displayID)
	if err != nil {
		return nil, fmt.Errorf("install ime: %w", err)
	}
	s.lock, s.notice, s.history, s.ime = locks, notices, recent, keyboard
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

func (s *simulator) terminate(pid int) server.Frame {
	if pid == s.homeSession.PID() || !s.runtime.Terminate(pid) {
		return s.runtime.LastFrame(displayID)
	}
	if s.chat != nil && s.chatSession.PID() == pid {
		s.chat, s.chatSession = nil, server.Session{}
	}
	if s.settings != nil && s.settingsSession.PID() == pid {
		s.settings, s.settingsSession = nil, server.Session{}
	}
	return s.activateHome()
}
func (s *simulator) frame() server.Frame         { return s.runtime.LastFrame(displayID) }
func (s *simulator) splashFrame() server.Frame   { return s.splash }
func (s *simulator) foreground() server.WindowID { return s.runtime.Foreground(displayID) }
