package server

import (
	"testing"

	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

func appWindow() WindowConfig {
	return WindowConfig{Kind: AppWindow, Bounds: geom.Rect{Width: 320, Height: 480}}
}

func TestRuntimeLaunchesInstalledPackageAndRegistersProcess(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat-app", EntryPoint: "Chat.Main"})

	if _, err := runtime.Launch("missing-app"); err == nil {
		t.Fatal("uninstalled package must not launch")
	}

	launch, err := runtime.Launch("chat-app")
	if err != nil || launch.Session.PID() != 1 ||
		launch.Package.EntryPoint != "Chat.Main" || launch.Reused {
		t.Fatalf("launch = %#v, %v", launch, err)
	}
	if state, ok := runtime.ProcessState(launch.Session.PID()); !ok ||
		state != Starting {
		t.Fatalf("new process state = %v, found=%t; want Starting", state, ok)
	}
	if !runtime.processes.Has(launch.Session.PID()) ||
		runtime.processes.Len() != 1 {
		t.Fatal("ProcessRegistry must retain the process allocated by Runtime")
	}
	if !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Activate(launch.Session, "chat") {
		t.Fatal("prepare process window")
	}

	if !runtime.Terminate(launch.Session.PID()) {
		t.Fatal("system must be able to terminate a process")
	}
	if runtime.processes.Has(launch.Session.PID()) ||
		runtime.processes.Len() != 0 {
		t.Fatal("termination must remove the process from ProcessRegistry")
	}
	if runtime.Foreground(PrimaryDisplay) != "" ||
		runtime.Submit(launch.Session, "chat", model.Buffer{Bounds: geom.Rect{Width: 320, Height: 480}}) {
		t.Fatal("termination must remove the window and invalidate its session")
	}
}

func TestRuntimeRejectsRequestsForAnotherProcess(t *testing.T) {
	runtime := NewRuntime(
		&model.AppPackage{ID: "home", EntryPoint: "Home.Main"},
		&model.AppPackage{ID: "chat-app", EntryPoint: "Chat.Main"},
	)
	homeLaunch, err := runtime.Launch("home")
	if err != nil {
		t.Fatal(err)
	}
	chatLaunch, err := runtime.Launch("chat-app")
	if err != nil {
		t.Fatal(err)
	}
	home, chat := homeLaunch.Session, chatLaunch.Session
	if !runtime.AttachWindow(home, "home", PrimaryDisplay, appWindow()) {
		t.Fatal("home must attach its own window")
	}
	if runtime.AttachWindow(chat, "home", PrimaryDisplay, appWindow()) {
		t.Fatal("another process must not attach an existing window")
	}
	if runtime.Activate(chat, "home") {
		t.Fatal("another process must not activate a foreign window")
	}
	if runtime.Submit(chat, "home", model.Buffer{Content: "malicious"}) {
		t.Fatal("another process must not submit to a foreign window")
	}
	forged := Session{pid: home.PID()}
	if runtime.Submit(forged, "home", model.Buffer{Content: "forged"}) {
		t.Fatal("a PID without the server-issued credential must not be authorized")
	}
	if runtime.Terminate(999) {
		t.Fatal("unknown process must not terminate")
	}
	if runtime.Terminate(home.PID()) == false {
		t.Fatal("system must terminate the selected process")
	}
	if runtime.Submit(home, "home", model.Buffer{Content: "stale"}) {
		t.Fatal("terminated session must no longer be authorized")
	}
}

func TestRuntimeOwnsProcessState(t *testing.T) {
	runtime := NewRuntime(
		&model.AppPackage{ID: "home", EntryPoint: "Home.Main"},
		&model.AppPackage{ID: "chat-app", EntryPoint: "Chat.Main"},
	)
	homeLaunch, err := runtime.Launch("home")
	if err != nil {
		t.Fatal(err)
	}
	chatLaunch, err := runtime.Launch("chat-app")
	if err != nil {
		t.Fatal(err)
	}
	home, chat := homeLaunch.Session, chatLaunch.Session
	if !runtime.AttachWindow(home, "home", PrimaryDisplay, appWindow()) ||
		!runtime.AttachWindow(chat, "chat", PrimaryDisplay, appWindow()) {
		t.Fatal("attach application windows")
	}
	if !runtime.Activate(home, "home") {
		t.Fatal("activate home")
	}
	if runtime.processes.processes[home.PID()].State != Active ||
		runtime.processes.processes[chat.PID()].State != Cached {
		t.Fatal("server must mark only the foreground process active")
	}
	if !runtime.Activate(chat, "chat") {
		t.Fatal("activate chat")
	}
	if runtime.processes.processes[home.PID()].State != Cached ||
		runtime.processes.processes[chat.PID()].State != Active {
		t.Fatal("server must update process state when foreground changes")
	}
}

func TestRuntimeLaunchResumesCachedProcessAndRestartsEvictedProcess(
	t *testing.T,
) {
	runtime := NewRuntime(
		&model.AppPackage{ID: "home", EntryPoint: "Home.Main"},
		&model.AppPackage{ID: "chat-app", EntryPoint: "Chat.Main"},
	)
	home, err := runtime.Launch("home")
	if err != nil {
		t.Fatal(err)
	}
	chat, err := runtime.Launch("chat-app")
	if err != nil {
		t.Fatal(err)
	}
	if !runtime.AttachWindow(home.Session, "home", PrimaryDisplay, appWindow()) ||
		!runtime.AttachWindow(chat.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Activate(chat.Session, "chat") ||
		!runtime.Activate(home.Session, "home") {
		t.Fatal("prepare cached chat process")
	}
	if state, _ := runtime.ProcessState(chat.Session.PID()); state != Cached {
		t.Fatalf("chat state = %v, want Cached", state)
	}

	resumed, err := runtime.Launch("chat-app")
	if err != nil || !resumed.Reused ||
		resumed.Session.PID() != chat.Session.PID() {
		t.Fatalf("warm launch = %#v, %v", resumed, err)
	}
	if !runtime.Evict(chat.Session.PID()) {
		t.Fatal("cached process must be evictable")
	}
	if state, _ := runtime.ProcessState(chat.Session.PID()); state != Evicted {
		t.Fatalf("evicted state = %v, want Evicted", state)
	}
	if runtime.Submit(chat.Session, "chat", model.Buffer{}) {
		t.Fatal("eviction must invalidate the old session")
	}

	restarted, err := runtime.Launch("chat-app")
	if err != nil || restarted.Reused ||
		restarted.Session.PID() == chat.Session.PID() {
		t.Fatalf("cold relaunch = %#v, %v", restarted, err)
	}
}

func TestRuntimePresentConfirmsComposedBuffer(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat-app", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat-app")
	if err != nil {
		t.Fatal(err)
	}
	if !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Activate(launch.Session, "chat") {
		t.Fatal("prepare chat window")
	}
	buffer := model.Buffer{
		Revision: 1,
		Content:  "first frame",
		Bounds:   displayBounds(t, &runtime, PrimaryDisplay),
	}
	frame, presented := runtime.Present(launch.Session, "chat", buffer)
	if !presented || !frame.Presents("chat", buffer.Revision) {
		t.Fatalf("first frame = %#v, presented=%t", frame, presented)
	}

	if _, presented := runtime.Present(
		Session{},
		"chat",
		model.Buffer{Revision: 2},
	); presented {
		t.Fatal("unauthorized submission must not report a presented frame")
	}
}

func TestRuntimeRotateRequiresResizedBuffer(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat-app", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat-app")
	if err != nil {
		t.Fatal(err)
	}
	if !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Activate(launch.Session, "chat") {
		t.Fatal("prepare chat window")
	}
	oldBounds := displayBounds(t, &runtime, PrimaryDisplay)
	display, ok := runtime.Rotate(PrimaryDisplay)
	if !ok {
		t.Fatal("rotate primary display")
	}
	if display.Orientation != RightUp || display.Width != 480 ||
		display.Height != 320 || display.Revision != 2 {
		t.Fatalf("rotated display = %#v", display)
	}
	if _, presented := runtime.Present(launch.Session, "chat", model.Buffer{
		Revision: 1,
		Bounds:   oldBounds,
	}); presented {
		t.Fatal("buffer with pre-rotation bounds must not be presented")
	}
	if !runtime.ResizeWindow(launch.Session, "chat", display.Bounds()) {
		t.Fatal("app owner must resize its window for the rotated display")
	}
	frame, presented := runtime.Present(launch.Session, "chat", model.Buffer{
		Revision: 2,
		Bounds:   display.Bounds(),
	})
	if !presented || frame.Display != display {
		t.Fatalf("rotated frame = %#v, presented=%t", frame, presented)
	}
}

func TestDisplayConfigRotateCyclesAllOrientations(t *testing.T) {
	config := DisplayConfig{
		Width: 320, Height: 480, Orientation: TopUp, Revision: 1,
	}
	want := []DisplayConfig{
		{Width: 480, Height: 320, Orientation: RightUp, Revision: 2},
		{Width: 320, Height: 480, Orientation: BottomUp, Revision: 3},
		{Width: 480, Height: 320, Orientation: LeftUp, Revision: 4},
		{Width: 320, Height: 480, Orientation: TopUp, Revision: 5},
	}
	for _, next := range want {
		config = config.Rotate()
		if config != next {
			t.Fatalf("rotated config = %#v, want %#v", config, next)
		}
	}
}

func TestDisplayEdgeAliases(t *testing.T) {
	if Portrait != TopUp || Landscape != RightUp {
		t.Fatalf("aliases = portrait:%v landscape:%v", Portrait, Landscape)
	}
}

func TestRuntimeConfigAdvancesRevisionAndCanLockOrientation(t *testing.T) {
	runtime := NewRuntime()
	config := runtime.Config()
	config.Theme = model.DarkTheme
	config.FontScale = 1.25
	config.OrientationLocked = true
	updated := runtime.SetConfig(config)
	if updated.Revision != 2 || updated.Theme != model.DarkTheme ||
		updated.FontScale != 1.25 {
		t.Fatalf("config = %#v", updated)
	}
	if _, rotated := runtime.Rotate(PrimaryDisplay); rotated {
		t.Fatal("orientation lock must reject rotation")
	}
}

func TestRuntimeMarksANRAndCrashSeparately(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil || !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Activate(launch.Session, "chat") {
		t.Fatal("prepare active process")
	}
	if !runtime.MarkUnresponsive(launch.Session.PID()) {
		t.Fatal("mark ANR")
	}
	if state, _ := runtime.ProcessState(launch.Session.PID()); state != Unresponsive {
		t.Fatalf("ANR state = %v", state)
	}
	if !runtime.Crash(launch.Session.PID()) {
		t.Fatal("crash unresponsive process")
	}
	if state, _ := runtime.ProcessState(launch.Session.PID()); state != Crashed {
		t.Fatalf("crash state = %v", state)
	}
	if runtime.Submit(launch.Session, "chat", model.Buffer{}) {
		t.Fatal("crashed session must be invalid")
	}
}

func TestRuntimeCaptureExcludesProtectedWindows(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil || !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Activate(launch.Session, "chat") {
		t.Fatal("prepare app")
	}
	if !runtime.RegisterSystemWindow("lock", PrimaryDisplay, displayBounds(t, &runtime, PrimaryDisplay)) {
		t.Fatal("register protected window")
	}
	if !runtime.ProtectWindow("lock") {
		t.Fatal("protect lock window")
	}
	if !runtime.Submit(launch.Session, "chat", model.Buffer{Revision: 1, Bounds: displayBounds(t, &runtime, PrimaryDisplay)}) ||
		!runtime.SubmitSystem("lock", model.Buffer{Revision: 1, Bounds: displayBounds(t, &runtime, PrimaryDisplay)}) {
		t.Fatal("submit buffers")
	}
	runtime.VSync(PrimaryDisplay)
	capture, ok := runtime.Capture(PrimaryDisplay)
	if !ok || !capture.Presents("chat", 1) || capture.Presents("lock", 1) {
		t.Fatalf("capture = %#v found=%t", capture, ok)
	}
}

func TestRuntimeSplitShowsTwoPanesAndRoutesToTouchedPane(t *testing.T) {
	runtime := NewRuntime(
		&model.AppPackage{ID: "home", EntryPoint: "Home.Main"},
		&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"},
	)
	home, err := runtime.Launch("home")
	if err != nil {
		t.Fatal(err)
	}
	chat, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	if !runtime.AttachWindow(home.Session, "home", PrimaryDisplay, appWindow()) ||
		!runtime.AttachWindow(chat.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Split(PrimaryDisplay, "home", "chat") {
		t.Fatal("prepare split view")
	}
	frame := runtime.Compose(PrimaryDisplay)
	if runtime.Foreground(PrimaryDisplay) != "home" || len(frame.Layers) != 2 ||
		frame.Layers[0].WindowID != "home" || frame.Layers[1].WindowID != "chat" {
		t.Fatalf("split foreground=%q frame=%#v", runtime.Foreground(PrimaryDisplay), frame)
	}
	routed, ok := runtime.RouteInput(PrimaryDisplay, model.PointerEvent{
		Action:           model.PointerDown,
		ChangedPointerID: 0,
		Pointers:         []model.PointerSample{{ID: 0, X: 240, Y: 10}},
	})
	if !ok || routed.WindowID != "chat" || runtime.Foreground(PrimaryDisplay) != "chat" {
		t.Fatalf("split route=%#v found=%t foreground=%q", routed, ok, runtime.Foreground(PrimaryDisplay))
	}
}

func TestRuntimeKeepsDisplayScenesIndependent(t *testing.T) {
	runtime := NewRuntime(
		&model.AppPackage{ID: "primary-app", EntryPoint: "Primary.Main"},
		&model.AppPackage{ID: "external-app", EntryPoint: "External.Main"},
	)
	external := DisplayID("external")
	if !runtime.AddDisplay(external, DisplayConfig{
		Width: 1920, Height: 1080, Orientation: TopUp, Revision: 1,
	}) {
		t.Fatal("add external display")
	}
	primary, _ := runtime.Launch("primary-app")
	secondary, _ := runtime.Launch("external-app")
	if !runtime.AttachWindow(primary.Session, "primary", PrimaryDisplay, appWindow()) ||
		!runtime.AttachWindow(secondary.Session, "external", external, WindowConfig{
			Kind: AppWindow, Bounds: geom.Rect{Width: 1920, Height: 1080},
		}) ||
		!runtime.Activate(primary.Session, "primary") ||
		!runtime.Activate(secondary.Session, "external") {
		t.Fatal("prepare independent display scenes")
	}
	if runtime.Foreground(PrimaryDisplay) != "primary" ||
		runtime.Foreground(external) != "external" {
		t.Fatal("each display must retain its own foreground window")
	}
	primaryFrame, presented := runtime.Present(primary.Session, "primary", model.Buffer{
		Revision: 1, Bounds: geom.Rect{Width: 320, Height: 480},
	})
	if !presented || primaryFrame.DisplayID != PrimaryDisplay ||
		!primaryFrame.Presents("primary", 1) || primaryFrame.Presents("external", 1) {
		t.Fatalf("primary frame = %#v, presented=%t", primaryFrame, presented)
	}
	externalFrame, presented := runtime.Present(secondary.Session, "external", model.Buffer{
		Revision: 1, Bounds: geom.Rect{Width: 1920, Height: 1080},
	})
	if !presented || externalFrame.DisplayID != external ||
		!externalFrame.Presents("external", 1) || externalFrame.Presents("primary", 1) {
		t.Fatalf("external frame = %#v, presented=%t", externalFrame, presented)
	}
	rotated, ok := runtime.Rotate(external)
	if !ok || rotated.Bounds() != (geom.Rect{Width: 1080, Height: 1920}) {
		t.Fatalf("external rotation = %#v, found=%t", rotated, ok)
	}
	primaryConfig, ok := runtime.Display(PrimaryDisplay)
	if !ok || primaryConfig.Bounds() != (geom.Rect{Width: 320, Height: 480}) {
		t.Fatalf("primary display changed = %#v, found=%t", primaryConfig, ok)
	}
}

func displayBounds(t *testing.T, runtime *Runtime, id DisplayID) geom.Rect {
	t.Helper()
	display, ok := runtime.Display(id)
	if !ok {
		t.Fatalf("display %q not found", id)
	}
	return display.Bounds()
}
