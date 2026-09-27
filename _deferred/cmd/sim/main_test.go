package main

import (
	"testing"

	"mmos/internal/server"
)

func TestSimulator(t *testing.T) {
	s, err := NewSimulator()
	if err != nil {
		t.Fatal(err)
	}
	if !s.SplashFrame().Presents(splashWindow, 1) {
		t.Fatalf("home splash = %#v", s.SplashFrame())
	}
	frame, changed := s.Input(tap(160, 210))
	if !s.SplashFrame().Presents(splashWindow, 1) {
		t.Fatalf("chat splash = %#v", s.SplashFrame())
	}
	if !changed || s.Foreground() != "chat" || len(frame.Layers) != 3 {
		t.Fatalf(
			"launch = changed:%t foreground:%q frame:%#v",
			changed,
			s.Foreground(),
			frame,
		)
	}
	launchFrame := frame.Number
	frame, changed = s.Input(tap(160, 236))
	if !changed || frame.Number != launchFrame+1 {
		t.Fatalf("tap = changed:%t frame:%#v", changed, frame)
	}
	frame, changed = s.Input(server.SystemEvent{Action: server.SystemNotices})
	if !changed || !frame.Presents(noticeWindow, frame.Display.Revision) || len(s.runtime.Notices()) != 1 {
		t.Fatalf("notices = changed:%t frame:%#v notices:%#v", changed, frame, s.runtime.Notices())
	}
	if _, changed = s.Input(tap(160, 100)); !changed || s.Foreground() != "chat" {
		t.Fatalf("notice action = changed:%t foreground:%q", changed, s.Foreground())
	}
	frame = s.SplitChat()
	if len(frame.Layers) != 3 || s.homeWindow().Bounds().Width != 160 ||
		s.chat.Process.Windows["chat"].Bounds().Width != 160 {
		t.Fatalf("split = %#v", frame)
	}
	if _, changed = s.Input(tap(240, 236)); s.Foreground() != "chat" {
		t.Fatalf("split focus = changed:%t foreground:%q", changed, s.Foreground())
	}
	frame, changed = s.Input(server.RotateEvent{})
	if !changed || frame.Display.Orientation != server.RightUp {
		t.Fatalf("rotate = changed:%t frame:%#v", changed, frame)
	}
	frame, changed = s.Input(server.SystemEvent{Action: server.SystemBack})
	if !changed || s.Foreground() != "home" {
		t.Fatalf("back = changed:%t foreground:%q frame:%#v", changed, s.Foreground(), frame)
	}
	frame, changed = s.Input(server.SystemEvent{Action: server.SystemRecents})
	if !changed || !frame.Presents(historyWindow, frame.Display.Revision) {
		t.Fatalf("history = changed:%t frame:%#v", changed, frame)
	}
	frame, changed = s.Input(tap(160, 210))
	if !changed || s.Foreground() != "chat" {
		t.Fatalf("history select = changed:%t foreground:%q", changed, s.Foreground())
	}
	frame, changed = s.Input(server.SystemEvent{Action: server.SystemLock})
	if !changed || !frame.Presents(lockWindow, frame.Display.Revision) {
		t.Fatalf("lock = changed:%t frame:%#v", changed, frame)
	}
	frame, changed = s.Input(server.SystemEvent{Action: server.SystemScreenOff})
	if !changed || frame.Display.Power != server.ScreenOff || len(frame.Layers) != 0 {
		t.Fatalf("screen off = changed:%t frame:%#v", changed, frame)
	}
	frame, changed = s.Input(server.SystemEvent{Action: server.SystemScreenOn})
	if !changed || frame.Display.Power != server.ScreenOn || !frame.Presents(lockWindow, frame.Display.Revision) {
		t.Fatalf("screen on = changed:%t frame:%#v", changed, frame)
	}
	if _, changed = s.Input(tap(160, 236)); changed {
		t.Fatal("locked simulator must reject app input")
	}
	frame, changed = s.Input(server.SystemEvent{Action: server.SystemUnlock})
	if !changed || frame.Presents(lockWindow, 1) {
		t.Fatalf("unlock = changed:%t frame:%#v", changed, frame)
	}
	frame, changed = s.Input(server.SystemEvent{Action: server.SystemSettings})
	if !changed || !frame.Presents(settingsWindow, frame.Display.Revision) {
		t.Fatalf("settings = changed:%t frame:%#v", changed, frame)
	}
	if _, changed = s.Input(tap(160, 10)); !changed ||
		s.runtime.Config().Theme != server.DarkTheme ||
		s.chat.Process.Config.Revision != s.runtime.Config().Revision {
		t.Fatalf("settings change = changed:%t config:%#v", changed, s.runtime.Config())
	}
	if _, changed = s.Input(tap(160, 230)); !changed ||
		s.runtime.Config().Locale != "ko-KR" {
		t.Fatalf("locale setting = changed:%t config:%#v", changed, s.runtime.Config())
	}
	if _, changed = s.Input(tap(160, 300)); !changed ||
		!s.runtime.Config().OrientationLocked {
		t.Fatalf("orientation lock setting = changed:%t config:%#v", changed, s.runtime.Config())
	}
	if _, changed = s.Input(server.SystemEvent{Action: server.SystemBack}); !changed {
		t.Fatal("back must close settings")
	}
	frame, changed = s.Input(server.SystemEvent{Action: server.SystemHome})
	if !changed || s.Foreground() != "home" {
		t.Fatalf("home = changed:%t foreground:%q frame:%#v", changed, s.Foreground(), frame)
	}
	if frame = s.Terminate(s.chatSession.PID()); s.Foreground() != "home" ||
		len(frame.Layers) != 2 {
		t.Fatalf("terminate = %#v", frame)
	}
}
