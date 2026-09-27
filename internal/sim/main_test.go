package sim

import (
	"testing"

	"mmos/internal/server"
)

func TestSimulatorCoreFlow(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, changed := tap(s, 160, 210); !changed || s.Foreground() != "chat" {
		t.Fatal("launch chat")
	}
	if _, changed := tap(s, 160, 236); !changed {
		t.Fatal("send message")
	}
	if frame := s.Back(); s.Foreground() != "home" || !frame.Presents("home", 2) {
		t.Fatalf("back = foreground:%q frame:%#v", s.Foreground(), frame)
	}
}

func TestSimulatorOpensNotificationShadeFromTopPull(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, changed := tap(s, 160, 210); !changed {
		t.Fatal("launch chat")
	}
	if _, changed := tap(s, 160, 236); !changed {
		t.Fatal("publish notice")
	}
	s.Input(pointer(server.PointerDown, 160, 8))
	frame, changed := s.Input(pointer(server.PointerUp, 160, 140))
	if !changed || !frame.Presents("notices", 1) {
		t.Fatalf("notification shade = %#v changed=%t", frame, changed)
	}
	for _, layer := range frame.Layers {
		if layer.WindowID == "notices" && layer.Buffer.Content != "notices(count=1, latest=Chat: Message sent)" {
			t.Fatalf("notice content = %q", layer.Buffer.Content)
		}
	}
}

func TestSimulatorShowsRecentsAndSelectsTask(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, changed := tap(s, 160, 210); !changed {
		t.Fatal("launch chat")
	}
	frame, changed := s.Input(server.SystemEvent{Action: server.SystemRecents})
	if !changed || !frame.Presents("history", 1) {
		t.Fatalf("recents: changed=%t frame=%#v", changed, frame)
	}
	frame, changed = tap(s, 160, 176)
	if !changed || s.Foreground() != "home" || !frame.Presents("home", 2) {
		t.Fatalf("select home: changed=%t foreground=%q frame=%#v", changed, s.Foreground(), frame)
	}
}

func pointer(action server.PointerAction, x, y int) server.PointerEvent {
	return server.PointerEvent{
		Source:           server.TouchSource,
		Action:           action,
		ChangedPointerID: 0,
		Pointers:         []server.PointerSample{{ID: 0, X: x, Y: y}},
	}
}

func TestSimulatorShowsSplashUntilAppFirstFrame(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if !s.SplashFrame().Presents("splash", 1) || s.Frame().Presents("splash", 1) {
		t.Fatalf("splash=%#v frame=%#v", s.SplashFrame(), s.Frame())
	}
}

func TestSimulatorRelayoutsClientsAfterRotation(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, changed := tap(s, 160, 210); !changed {
		t.Fatal("launch chat")
	}
	frame, changed := s.Input(server.RotateEvent{})
	if !changed || frame.Display.Bounds().Width != 480 || frame.Display.Bounds().Height != 320 ||
		s.homeWindow().Bounds() != frame.Display.Bounds() || s.chat.Process.Windows["chat"].Bounds() != frame.Display.Bounds() ||
		!frame.Presents("chat", 1) {
		t.Fatalf("rotation = changed:%t frame:%#v", changed, frame)
	}
}

func TestSimulatorKeepsVisibleWindowsInOneFrame(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, changed := tap(s, 160, 210); !changed {
		t.Fatal("launch chat")
	}
	frame := s.Frame()
	if len(frame.Layers) != 3 || !frame.Presents("status", 1) || !frame.Presents("home", 1) || !frame.Presents("chat", 1) {
		t.Fatalf("frame = %#v", frame)
	}
}

func TestSimulatorUnlocksThroughLockScreenControl(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	config := s.runtime.Config()
	config.LockOnScreenOff = true
	s.runtime.SetConfig(config)
	if _, changed := s.Input(server.SystemEvent{Action: server.SystemScreenOff}); !changed {
		t.Fatal("screen off")
	}
	if _, changed := s.Input(server.SystemEvent{Action: server.SystemScreenOn}); !changed {
		t.Fatal("screen on")
	}
	if _, changed := tap(s, 0, 0); changed {
		t.Fatal("unexpected unlock")
	}
	if _, changed := tap(s, 160, 432); !changed || s.Frame().Presents("lock", 1) {
		t.Fatalf("unlock = changed:%t frame:%#v", changed, s.Frame())
	}
}

func tap(s *Simulator, x, y int) (server.Frame, bool) {
	s.Input(pointer(server.PointerDown, x, y))
	return s.Input(pointer(server.PointerUp, x, y))
}
