package main

import (
	"fmt"

	"mmos/internal/app/chat"
	"mmos/internal/app/home"
	"mmos/internal/client"
	"mmos/internal/server"
)

const (
	statusWindow   server.WindowID = "status"
	splashWindow   server.WindowID = "splash"
	historyWindow  server.WindowID = "history"
	lockWindow     server.WindowID = "lock"
	settingsWindow server.WindowID = "settings"
	noticeWindow   server.WindowID = "notices"
)

const displayID = server.PrimaryDisplay

// Simulator composes the server and client applications for this demo.
type Simulator struct {
	runtime     server.Runtime
	drawer      client.Drawer
	homeApp     *home.App
	homeSession server.Session
	chat        *chat.App
	chatSession server.Session
	splash      server.Frame
	history     bool
	locked      bool
	settings    bool
	notices     bool
	split       bool
}

func NewSimulator() (*Simulator, error) {
	homePackage := &server.AppPackage{
		ID:         "home",
		EntryPoint: "Home.Main",
		Assets:     map[string][]byte{},
	}
	chatPackage := &server.AppPackage{
		ID:         "chat-app",
		EntryPoint: "Chat.Main",
		Assets:     map[string][]byte{},
	}
	runtime := server.NewRuntime(homePackage, chatPackage)
	homeLaunch, err := runtime.Launch("home")
	if err != nil {
		return nil, fmt.Errorf("launch home: %w", err)
	}
	s := &Simulator{runtime: runtime, homeSession: homeLaunch.Session}
	s.homeApp = home.New(homeLaunch.Package, s.LaunchChat)
	s.applyConfig(runtime.Config())
	statusBounds := server.Rect{Width: 320, Height: 24}
	if !s.runtime.RegisterSystemWindow(statusWindow, displayID, statusBounds) {
		return nil, fmt.Errorf("register status window")
	}
	if !s.runtime.SubmitSystem(statusWindow, server.Buffer{
		Revision: 1,
		Content:  "status",
		Bounds:   statusBounds,
	}) {
		return nil, fmt.Errorf("register status window")
	}
	if !s.registerOverlay(historyWindow) || !s.registerOverlay(lockWindow) ||
		!s.registerOverlay(settingsWindow) || !s.registerOverlay(noticeWindow) {
		return nil, fmt.Errorf("register system overlay")
	}
	if !s.runtime.ProtectWindow(lockWindow) {
		return nil, fmt.Errorf("protect lock window")
	}
	if !s.registerSplash() || !s.showSplash(homeLaunch.Package.ID) {
		return nil, fmt.Errorf("show home splash")
	}
	window := s.homeWindow()
	config := server.WindowConfig{
		Kind:   server.AppWindow,
		Bounds: window.Bounds(),
	}
	if !s.runtime.AttachWindow(s.homeSession, window.ID, displayID, config) ||
		!s.runtime.Activate(s.homeSession, window.ID) ||
		!s.draw(s.homeSession, s.homeApp.Process, window.ID) ||
		!s.runtime.HideSystemWindow(splashWindow) {
		return nil, fmt.Errorf("start home window")
	}
	return s, nil
}

func (s *Simulator) LaunchChat() bool {
	launch, err := s.runtime.Launch("chat-app")
	if err != nil {
		return false
	}
	if !s.showSplash(launch.Package.ID) {
		return false
	}
	splashVisible := true
	defer func() {
		if splashVisible {
			s.runtime.HideSystemWindow(splashWindow)
		}
	}()
	if !launch.Reused {
		s.chatSession = launch.Session
		s.chat = chat.New(launch.Package)
		s.chat.Process.Configure(s.runtime.Config())
		window := s.chat.Process.Windows["chat"]
		display, ok := s.runtime.Display(displayID)
		if !ok {
			return false
		}
		window.Resize(display.Bounds())
		config := server.WindowConfig{
			Kind:   server.AppWindow,
			Bounds: window.Bounds(),
		}
		if !s.runtime.AttachWindow(s.chatSession, window.ID, displayID, config) {
			s.runtime.HideSystemWindow(splashWindow)
			s.abortChat()
			return false
		}
	} else if s.chat == nil {
		return false
	} else {
		s.chatSession = launch.Session
	}
	if !s.runtime.TransitionTo(s.chatSession, "chat") {
		s.runtime.HideSystemWindow(splashWindow)
		return false
	}
	s.split = false
	if !s.runtime.HideSystemWindow(splashWindow) {
		return false
	}
	splashVisible = false
	_, presented := s.present(s.chatSession, s.chat.Process, "chat")
	if !presented && !launch.Reused {
		s.abortChat()
	}
	return presented
}

// Input accepts a normalized system input event. App-directed input is routed
// to the foreground window; system and display events are handled by their
// respective policies instead.
func (s *Simulator) Input(event server.InputEvent) (server.Frame, bool) {
	switch event := event.(type) {
	case server.SystemEvent:
		return s.systemEvent(event)
	case server.RotateEvent:
		return s.rotate(), true
	}
	if s.locked {
		return s.runtime.LastFrame(displayID), false
	}
	if s.history {
		return s.historyInput(event)
	}
	if s.settings {
		return s.settingsInput(event)
	}
	if s.notices {
		return s.noticeInput(event)
	}
	routed, ok := s.runtime.RouteInput(displayID, event)
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	if routed.WindowID == s.homeWindow().ID {
		if !s.homeApp.HandleInput(routed.WindowID, routed.Event) {
			return s.runtime.LastFrame(displayID), false
		}
		return s.runtime.LastFrame(displayID), true
	}
	if s.chat == nil || routed.WindowID != "chat" ||
		!s.chat.HandleInput(routed.WindowID, routed.Event) {
		return s.runtime.LastFrame(displayID), false
	}
	s.runtime.PublishNotice(s.chatSession, "Chat", "Message sent")
	return s.present(s.chatSession, s.chat.Process, routed.WindowID)
}

// Home is a convenience entry point for the simulation controls. Home still
// enters the system as a SystemEvent.
func (s *Simulator) Home() server.Frame {
	frame, _ := s.Input(server.SystemEvent{Action: server.SystemHome})
	return frame
}

// Rotate is a convenience entry point for the simulation controls. Rotation
// still enters the system as a RotateEvent.
func (s *Simulator) Rotate() server.Frame {
	frame, _ := s.Input(server.RotateEvent{})
	return frame
}

// SplitChat places Home and Chat side by side. The system assigns the pane
// bounds; each client only relays out and redraws its own window.
func (s *Simulator) SplitChat() server.Frame {
	if s.chat == nil || !s.runtime.Split(displayID, s.homeWindow().ID, "chat") {
		return s.runtime.LastFrame(displayID)
	}
	s.split = true
	display, ok := s.runtime.Display(displayID)
	if !ok {
		return s.runtime.LastFrame(displayID)
	}
	left, right := splitBounds(display.Bounds())
	if !s.resize(s.homeSession, s.homeApp.Process, s.homeWindow().ID, left) ||
		!s.resize(s.chatSession, s.chat.Process, "chat", right) ||
		!s.draw(s.homeSession, s.homeApp.Process, s.homeWindow().ID) ||
		!s.draw(s.chatSession, s.chat.Process, "chat") {
		return s.runtime.LastFrame(displayID)
	}
	return s.runtime.VSync(displayID)
}

// rotate applies display policy after a RotateEvent has been accepted.
func (s *Simulator) rotate() server.Frame {
	display, ok := s.runtime.Rotate(displayID)
	if !ok {
		return s.runtime.LastFrame(displayID)
	}
	status := server.Rect{Width: display.Width, Height: 24}
	statusBuffer := server.Buffer{
		Revision: display.Revision,
		Content:  "status",
		Bounds:   status,
	}
	if !s.runtime.ResizeSystemWindow(statusWindow, status) ||
		!s.runtime.SubmitSystem(statusWindow, statusBuffer) {
		return s.runtime.LastFrame(displayID)
	}
	homeBounds, chatBounds := display.Bounds(), display.Bounds()
	if s.split {
		homeBounds, chatBounds = splitBounds(display.Bounds())
	}
	if !s.resize(s.homeSession, s.homeApp.Process, s.homeWindow().ID, homeBounds) ||
		(s.chat != nil && !s.resize(s.chatSession, s.chat.Process, "chat", chatBounds)) {
		return s.runtime.LastFrame(displayID)
	}
	if s.history && !s.showOverlay(historyWindow, s.historyContent()) {
		return s.runtime.LastFrame(displayID)
	}
	if s.locked && !s.showOverlay(lockWindow, "lock") {
		return s.runtime.LastFrame(displayID)
	}
	if s.settings && !s.showOverlay(settingsWindow, s.settingsContent()) {
		return s.runtime.LastFrame(displayID)
	}
	if s.split {
		if !s.draw(s.homeSession, s.homeApp.Process, s.homeWindow().ID) ||
			!s.draw(s.chatSession, s.chat.Process, "chat") {
			return s.runtime.LastFrame(displayID)
		}
		return s.runtime.VSync(displayID)
	}
	if s.runtime.Foreground(displayID) == "chat" && s.chat != nil {
		frame, _ := s.present(s.chatSession, s.chat.Process, "chat")
		return frame
	}
	frame, _ := s.present(s.homeSession, s.homeApp.Process, s.homeWindow().ID)
	return frame
}

// Terminate asks the trusted runtime to end a process, then discards this
// harness's client instance for that process. Home remains the shell process
// in this single-window simulation and is not terminated by this control.
func (s *Simulator) Terminate(pid int) server.Frame {
	if pid == s.homeSession.PID() || !s.runtime.Terminate(pid) {
		return s.runtime.LastFrame(displayID)
	}
	if s.chat != nil && s.chatSession.PID() == pid {
		s.chat, s.chatSession = nil, server.Session{}
		s.split = false
	}
	return s.activateHome()
}

func (s *Simulator) Frame() server.Frame { return s.runtime.LastFrame(displayID) }

// SplashFrame reports the most recently composited launch splash.
func (s *Simulator) SplashFrame() server.Frame { return s.splash }

func (s *Simulator) Foreground() server.WindowID {
	return s.runtime.Foreground(displayID)
}

func (s *Simulator) systemEvent(event server.SystemEvent) (server.Frame, bool) {
	if s.locked && event.Action != server.SystemUnlock && event.Action != server.SystemScreenOn &&
		event.Action != server.SystemScreenOff {
		return s.runtime.LastFrame(displayID), false
	}
	switch event.Action {
	case server.SystemBack:
		if s.history {
			return s.hideHistory(), true
		}
		if s.settings {
			return s.hideSettings(), true
		}
		if s.chat != nil && s.runtime.Foreground(displayID) == "chat" && s.chat.Back() {
			frame, presented := s.present(s.chatSession, s.chat.Process, "chat")
			return frame, presented
		}
		return s.activateHome(), true
	case server.SystemHome:
		return s.activateHome(), true
	case server.SystemRecents:
		if s.history {
			return s.hideHistory(), true
		}
		return s.showHistory(), true
	case server.SystemLock:
		return s.showLock(), true
	case server.SystemUnlock:
		return s.unlock(), true
	case server.SystemSettings:
		if s.settings {
			return s.hideSettings(), true
		}
		return s.showSettings(), true
	case server.SystemNotices:
		if s.notices {
			return s.hideNotices(), true
		}
		return s.showNotices(), true
	case server.SystemScreenOff:
		return s.screenOff(), true
	case server.SystemScreenOn:
		return s.screenOn(), true
	default:
		return s.runtime.LastFrame(displayID), false
	}
}

func (s *Simulator) historyInput(event server.InputEvent) (server.Frame, bool) {
	pointer, ok := event.(server.PointerEvent)
	if !ok || pointer.Action != server.PointerUp || pointer.ChangedPointerID != 0 {
		return s.runtime.LastFrame(displayID), false
	}
	if s.chat != nil {
		if !s.runtime.HideSystemWindow(historyWindow) ||
			!s.runtime.TransitionTo(s.chatSession, "chat") {
			return s.runtime.LastFrame(displayID), false
		}
		s.history = false
		frame, presented := s.present(s.chatSession, s.chat.Process, "chat")
		return frame, presented
	}
	return s.hideHistory(), true
}

func (s *Simulator) showHistory() server.Frame {
	if !s.showOverlay(historyWindow, s.historyContent()) {
		return s.runtime.LastFrame(displayID)
	}
	s.history = true
	return s.runtime.VSync(displayID)
}

func (s *Simulator) hideHistory() server.Frame {
	if !s.runtime.HideSystemWindow(historyWindow) {
		return s.runtime.LastFrame(displayID)
	}
	s.history = false
	return s.runtime.VSync(displayID)
}

func (s *Simulator) showLock() server.Frame {
	if !s.showOverlay(lockWindow, "lock") {
		return s.runtime.LastFrame(displayID)
	}
	s.locked = true
	return s.runtime.VSync(displayID)
}

func (s *Simulator) unlock() server.Frame {
	display, ok := s.runtime.Display(displayID)
	if !ok || display.Power == server.ScreenOff {
		return s.runtime.LastFrame(displayID)
	}
	if !s.runtime.HideSystemWindow(lockWindow) {
		return s.runtime.LastFrame(displayID)
	}
	s.locked = false
	return s.runtime.VSync(displayID)
}

func (s *Simulator) screenOff() server.Frame {
	s.showLock()
	if _, ok := s.runtime.SetScreenPower(displayID, server.ScreenOff); !ok {
		return s.runtime.LastFrame(displayID)
	}
	return s.runtime.VSync(displayID)
}

func (s *Simulator) screenOn() server.Frame {
	if _, ok := s.runtime.SetScreenPower(displayID, server.ScreenOn); !ok {
		return s.runtime.LastFrame(displayID)
	}
	return s.runtime.VSync(displayID)
}

func (s *Simulator) settingsInput(event server.InputEvent) (server.Frame, bool) {
	pointer, ok := event.(server.PointerEvent)
	if !ok || pointer.Action != server.PointerUp || pointer.ChangedPointerID != 0 {
		return s.runtime.LastFrame(displayID), false
	}
	sample, ok := pointer.PointerByID(pointer.ChangedPointerID)
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	config := s.runtime.Config()
	display, ok := s.runtime.Display(displayID)
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	switch {
	case sample.Y < display.Height/5:
		if config.Theme == server.LightTheme {
			config.Theme = server.DarkTheme
		} else {
			config.Theme = server.LightTheme
		}
	case sample.Y < display.Height*2/5:
		config.FontScale += .25
		if config.FontScale > 1.5 {
			config.FontScale = 1
		}
	case sample.Y < display.Height*3/5:
		config.ReducedMotion = !config.ReducedMotion
	case sample.Y < display.Height*4/5:
		config.Locale = nextLocale(config.Locale)
	default:
		config.OrientationLocked = !config.OrientationLocked
	}
	s.applyConfig(s.runtime.SetConfig(config))
	if !s.showOverlay(settingsWindow, s.settingsContent()) {
		return s.runtime.LastFrame(displayID), false
	}
	return s.presentForeground()
}

func (s *Simulator) showSettings() server.Frame {
	if !s.showOverlay(settingsWindow, s.settingsContent()) {
		return s.runtime.LastFrame(displayID)
	}
	s.settings = true
	return s.runtime.VSync(displayID)
}

func (s *Simulator) hideSettings() server.Frame {
	if !s.runtime.HideSystemWindow(settingsWindow) {
		return s.runtime.LastFrame(displayID)
	}
	s.settings = false
	return s.runtime.VSync(displayID)
}

func (s *Simulator) noticeInput(event server.InputEvent) (server.Frame, bool) {
	pointer, ok := event.(server.PointerEvent)
	if !ok || pointer.Action != server.PointerUp {
		return s.runtime.LastFrame(displayID), false
	}
	if !s.runtime.HideSystemWindow(noticeWindow) {
		return s.runtime.LastFrame(displayID), false
	}
	s.notices = false
	if s.chat != nil {
		if !s.runtime.TransitionTo(s.chatSession, "chat") {
			return s.runtime.LastFrame(displayID), false
		}
		frame, presented := s.present(s.chatSession, s.chat.Process, "chat")
		return frame, presented
	}
	return s.runtime.VSync(displayID), true
}

func (s *Simulator) showNotices() server.Frame {
	items := s.runtime.Notices()
	if !s.showOverlay(noticeWindow, fmt.Sprintf("notices(count=%d)", len(items))) {
		return s.runtime.LastFrame(displayID)
	}
	s.notices = true
	return s.runtime.VSync(displayID)
}

func (s *Simulator) hideNotices() server.Frame {
	if !s.runtime.HideSystemWindow(noticeWindow) {
		return s.runtime.LastFrame(displayID)
	}
	s.notices = false
	return s.runtime.VSync(displayID)
}

func (s *Simulator) activateHome() server.Frame {
	if !s.runtime.TransitionTo(s.homeSession, s.homeWindow().ID) {
		return s.runtime.LastFrame(displayID)
	}
	s.split = false
	frame, ok := s.present(s.homeSession, s.homeApp.Process, s.homeWindow().ID)
	if !ok {
		return s.runtime.LastFrame(displayID)
	}
	return frame
}

func (s *Simulator) abortChat() {
	s.runtime.Terminate(s.chatSession.PID())
	s.chat, s.chatSession = nil, server.Session{}
}

func (s *Simulator) draw(session server.Session, process *client.AppProcess, id server.WindowID) bool {
	window, ok := process.Windows[id]
	if !ok {
		return false
	}

	s.drawer.Draw(window)
	return s.runtime.Submit(session, id, window.Buffer)
}

func (s *Simulator) present(
	session server.Session,
	process *client.AppProcess,
	id server.WindowID,
) (server.Frame, bool) {
	window, ok := process.Windows[id]
	if !ok {
		return s.runtime.LastFrame(displayID), false
	}
	s.drawer.Draw(window)
	return s.runtime.Present(session, id, window.Buffer)
}

func (s *Simulator) resize(
	session server.Session,
	process *client.AppProcess,
	id server.WindowID,
	bounds server.Rect,
) bool {
	window, ok := process.Windows[id]
	if !ok || !s.runtime.ResizeWindow(session, id, bounds) {
		return false
	}
	window.Resize(bounds)
	return true
}

func (s *Simulator) homeWindow() *client.Window {
	return s.homeApp.Process.Windows["home"]
}

func (s *Simulator) registerSplash() bool {
	display, ok := s.runtime.Display(displayID)
	if !ok || !s.runtime.RegisterSystemWindow(splashWindow, displayID, display.Bounds()) {
		return false
	}
	return s.runtime.HideSystemWindow(splashWindow)
}

func (s *Simulator) registerOverlay(id server.WindowID) bool {
	display, ok := s.runtime.Display(displayID)
	if !ok || !s.runtime.RegisterSystemWindow(id, displayID, display.Bounds()) {
		return false
	}
	return s.runtime.HideSystemWindow(id)
}

func (s *Simulator) showOverlay(id server.WindowID, content string) bool {
	display, ok := s.runtime.Display(displayID)
	if !ok || !s.runtime.ResizeSystemWindow(id, display.Bounds()) {
		return false
	}
	return s.runtime.SubmitSystem(id, server.Buffer{
		Revision: display.Revision,
		Content:  content,
		Bounds:   display.Bounds(),
	}) && s.runtime.ShowSystemWindow(id)
}

func (s *Simulator) historyContent() string {
	if s.chat == nil {
		return "history(home)"
	}
	return "history(home,chat)"
}

func splitBounds(bounds server.Rect) (server.Rect, server.Rect) {
	leftWidth := bounds.Width / 2
	return server.Rect{X: bounds.X, Y: bounds.Y, Width: leftWidth, Height: bounds.Height},
		server.Rect{X: bounds.X + leftWidth, Y: bounds.Y, Width: bounds.Width - leftWidth, Height: bounds.Height}
}

func (s *Simulator) settingsContent() string {
	config := s.runtime.Config()
	return fmt.Sprintf(
		"settings(theme=%d,font=%.2f,reduced-motion=%t,locale=%s,orientation-locked=%t)",
		config.Theme,
		config.FontScale,
		config.ReducedMotion,
		config.Locale,
		config.OrientationLocked,
	)
}

func nextLocale(locale string) string {
	switch locale {
	case "en-US":
		return "ko-KR"
	case "ko-KR":
		return "ja-JP"
	default:
		return "en-US"
	}
}

func (s *Simulator) applyConfig(config server.SystemConfig) {
	s.homeApp.Process.Configure(config)
	if s.chat != nil {
		s.chat.Process.Configure(config)
	}
}

func (s *Simulator) presentForeground() (server.Frame, bool) {
	if s.runtime.Foreground(displayID) == "chat" && s.chat != nil {
		return s.present(s.chatSession, s.chat.Process, "chat")
	}
	return s.present(s.homeSession, s.homeApp.Process, s.homeWindow().ID)
}

func (s *Simulator) showSplash(packageID string) bool {
	display, ok := s.runtime.Display(displayID)
	if !ok || !s.runtime.ResizeSystemWindow(splashWindow, display.Bounds()) {
		return false
	}
	buffer := server.Buffer{
		Revision: display.Revision,
		Content:  fmt.Sprintf("splash(package=%s)", packageID),
		Bounds:   display.Bounds(),
	}
	if !s.runtime.SubmitSystem(splashWindow, buffer) ||
		!s.runtime.ShowSystemWindow(splashWindow) {
		return false
	}
	s.splash = s.runtime.VSync(displayID)
	return s.splash.Presents(splashWindow, buffer.Revision)
}
