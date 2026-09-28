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

func TestSimulatorChangesConfigurationFromSettings(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, changed := tap(s, 160, 290); !changed || s.Foreground() != "settings" {
		t.Fatal("launch settings")
	}
	if _, changed := tap(s, 160, 120); !changed || s.runtime.Config().Theme != server.DarkTheme {
		t.Fatalf("toggle theme: changed=%t config=%#v", changed, s.runtime.Config())
	}
	if _, changed := tap(s, 160, 200); !changed || !s.runtime.Config().LockOnScreenOff {
		t.Fatalf("toggle lock: changed=%t config=%#v", changed, s.runtime.Config())
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

func TestSimulatorRoutesKeyboardCompositionToFocusedChat(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, changed := tap(s, 160, 210); !changed {
		t.Fatal("launch chat")
	}
	if _, changed := tap(s, 40, 150); !changed || !s.ime.Visible() {
		t.Fatal("focus editor and show ime")
	}
	if s.chat.State.KeyboardInset == 0 || s.chat.Process.Windows["chat"].Node("send").Bounds.Y >= 336 {
		t.Fatalf("chat did not move above ime: %#v", s.chat.State)
	}
	tap(s, 112, 360) // ㄱ
	tap(s, 240, 408) // ㅏ
	if s.chat.State.Composition != "가" || s.chat.State.Draft != "" {
		t.Fatalf("preedit = %#v", s.chat.State)
	}
	if _, changed := tap(s, 160, 456); !changed || s.chat.State.Draft != "가" || s.chat.State.Composition != "" {
		t.Fatalf("commit = %#v changed=%t", s.chat.State, changed)
	}
	if _, changed := s.Input(server.SystemEvent{Action: server.SystemBack}); !changed || s.ime.Visible() || s.chat.State.KeyboardInset != 0 {
		t.Fatalf("hide ime = visible:%t state:%#v", s.ime.Visible(), s.chat.State)
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
