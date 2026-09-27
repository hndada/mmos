package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

func TestDisplayContentBoundsAndCutoutRotate(t *testing.T) {
	config := DisplayConfig{
		Width: 320, Height: 480, Density: 2,
		SafeArea: geom.Insets{Top: 30, Bottom: 10},
		Cutout:   geom.Rect{X: 100, Y: 0, Width: 120, Height: 30},
	}
	if bounds := config.ContentBounds(); bounds != (geom.Rect{X: 0, Y: 30, Width: 320, Height: 440}) {
		t.Fatalf("content bounds = %#v", bounds)
	}
	rotated := config.Rotate()
	if rotated.SafeArea != (geom.Insets{Right: 30, Left: 10}) ||
		rotated.Cutout != (geom.Rect{X: 290, Y: 100, Width: 30, Height: 120}) {
		t.Fatalf("rotated display metrics = %#v", rotated)
	}
}

func TestInstallVerifiesSignatureAndPreventsDowngrade(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	runtime := NewRuntime()
	pkg := &model.AppPackage{ID: "signed", Version: 1, EntryPoint: "Signed.Main", Assets: map[string][]byte{"a": {1}}}
	signature := ed25519.Sign(private, PackageDigest(pkg))
	if err := runtime.Install(pkg, public, signature); err != nil {
		t.Fatalf("install: %v", err)
	}
	if _, err := runtime.Launch(pkg.ID); err != nil {
		t.Fatalf("launch signed package: %v", err)
	}
	if err := runtime.Install(pkg, public, signature); err != ErrPackageDowngrade {
		t.Fatalf("same version install = %v", err)
	}
}

func TestPermissionDecisionIsSystemOwned(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	request, state, ok := runtime.RequestPermission(launch.Session, Camera)
	if !ok || state != PermissionAsk || request.PackageID != "chat" {
		t.Fatalf("request = %#v state=%v ok=%t", request, state, ok)
	}
	if _, err := runtime.ResolvePermission(request.ID, true); err != nil {
		t.Fatal(err)
	}
	if state, ok := runtime.PermissionState(launch.Session, Camera); !ok || state != PermissionGranted {
		t.Fatalf("permission = %v found=%t", state, ok)
	}
}

func TestRecordingFiltersProtectedLayers(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	bounds := displayBounds(t, &runtime, PrimaryDisplay)
	if !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Activate(launch.Session, "chat") ||
		!runtime.RegisterSystemWindow("lock", PrimaryDisplay, bounds) ||
		!runtime.ProtectWindow("lock") ||
		!runtime.Submit(launch.Session, "chat", model.Buffer{Revision: 1, Bounds: bounds}) ||
		!runtime.SubmitSystem("lock", model.Buffer{Revision: 1, Bounds: bounds}) {
		t.Fatal("prepare recording")
	}
	if err := runtime.StartRecording(PrimaryDisplay); err != nil {
		t.Fatal(err)
	}
	runtime.VSync(PrimaryDisplay)
	recording, ok := runtime.StopRecording(PrimaryDisplay)
	if !ok || len(recording.Frames) != 1 || recording.Frames[0].Presents("lock", 1) {
		t.Fatalf("recording = %#v found=%t", recording, ok)
	}
}

func TestTasksAreSnapshots(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "one", EntryPoint: "One.Main"}, &model.AppPackage{ID: "two", EntryPoint: "Two.Main"})
	if _, err := runtime.Launch("two"); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Launch("one"); err != nil {
		t.Fatal(err)
	}
	tasks := runtime.Tasks()
	if len(tasks) != 2 || tasks[0].PID != 1 || tasks[1].PID != 2 || tasks[0].PackageID != "two" {
		t.Fatalf("tasks = %#v", tasks)
	}
}

func TestWatchdogMarksOnlyExpiredActiveProcess(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil || !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Activate(launch.Session, "chat") {
		t.Fatal("prepare active process")
	}
	watchdog := NewWatchdog(time.Second)
	now := time.Unix(10, 0)
	watchdog.Observe(launch.Session.PID(), now)
	if expired := runtime.CheckWatchdog(&watchdog, now.Add(999*time.Millisecond)); len(expired) != 0 {
		t.Fatalf("early expiry = %v", expired)
	}
	if expired := runtime.CheckWatchdog(&watchdog, now.Add(time.Second)); len(expired) != 1 || expired[0] != launch.Session.PID() {
		t.Fatalf("expiry = %v", expired)
	}
	if state, _ := runtime.ProcessState(launch.Session.PID()); state != Unresponsive {
		t.Fatalf("state = %v", state)
	}
}

func TestResourceLimitRejectsOversizeBuffersAndExcessWork(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil || !runtime.AttachWindow(launch.Session, "chat", PrimaryDisplay, appWindow()) ||
		!runtime.Activate(launch.Session, "chat") {
		t.Fatal("prepare active process")
	}
	if !runtime.SetResourceLimit("chat", ResourceLimit{MemoryBytes: 100, CPUUnits: 2, GPUUnits: 1}) {
		t.Fatal("set limit")
	}
	if runtime.Submit(launch.Session, "chat", model.Buffer{Bounds: geom.Rect{Width: 320, Height: 480}}) {
		t.Fatal("oversize buffer must be rejected")
	}
	if !runtime.Charge(launch.Session, 2, 1) || runtime.Charge(launch.Session, 1, 0) {
		t.Fatal("work quota must be enforced")
	}
}
