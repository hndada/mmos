package server

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"time"

	"mmos/internal/common/geom"
)

var ErrPackageNotInstalled = errors.New("package is not installed")
var ErrDisplayNotFound = errors.New("display not found")

// Launch is the result of starting or resuming an installed package.
type Launch struct {
	Session Session
	Package *AppPackage
	Reused  bool
	State   []byte
}

// Runtime owns the server-side process and display lifecycle for one system.
// Client applications create their own UI, then attach submitted windows to
// the process that the runtime started.
type Runtime struct {
	packages    map[string]*AppPackage
	signers     map[string][]byte
	permissions permissions
	limits      map[string]ResourceLimit
	usage       map[int]resourceUse
	protected   map[WindowID]bool
	recordings  map[DisplayID]Recording
	notices     []Notice
	nextNotice  int
	saved       map[string][]byte
	config      SystemConfig
	processes   ProcessRegistry
	displays    map[DisplayID]*display
	byProcess   map[int]map[WindowID]struct{}
	byPackage   map[string]int
	credentials map[int][32]byte
	owner       map[WindowID]int
	displayOf   map[WindowID]DisplayID
}

type display struct {
	config     DisplayConfig
	windows    WindowServer
	input      InputDispatcher
	compositor Compositor
	transition transition
}

func NewRuntime(installed ...*AppPackage) Runtime {
	runtime := Runtime{
		packages:    make(map[string]*AppPackage, len(installed)),
		signers:     map[string][]byte{},
		permissions: newPermissions(),
		limits:      map[string]ResourceLimit{},
		usage:       map[int]resourceUse{},
		protected:   map[WindowID]bool{},
		recordings:  map[DisplayID]Recording{},
		saved:       map[string][]byte{},
		config:      DefaultSystemConfig(),
		processes:   NewProcessRegistry(),
		displays: map[DisplayID]*display{
			PrimaryDisplay: {
				config:  DisplayConfig{Width: 320, Height: 480, Revision: 1, Density: 1},
				windows: NewWindowServer(),
				input:   NewInputDispatcher(),
			},
		},
		byProcess:   map[int]map[WindowID]struct{}{},
		byPackage:   map[string]int{},
		credentials: map[int][32]byte{},
		owner:       map[WindowID]int{},
		displayOf:   map[WindowID]DisplayID{},
	}
	for _, pkg := range installed {
		copy := *pkg
		copy.Assets = copyAssets(pkg.Assets)
		runtime.packages[pkg.ID] = &copy
	}
	return runtime
}

// Launch resumes a cached application when possible; otherwise it starts a
// new process. It is a trusted system operation.
func (r *Runtime) Launch(packageID string) (Launch, error) {
	pkg, ok := r.packages[packageID]
	if !ok {
		return Launch{}, ErrPackageNotInstalled
	}
	if pid, ok := r.byPackage[packageID]; ok {
		state, live := r.processes.State(pid)
		if live && (state == Active || state == Cached) {
			return Launch{Session: r.session(pid), Package: pkg, Reused: true, State: append([]byte(nil), r.saved[packageID]...)}, nil
		}
		delete(r.byPackage, packageID)
	}
	return r.start(pkg, true)
}

// LaunchInstance starts another process for packages that support multiple
// windows. Unlike Launch, it never resumes the package's primary process.
func (r *Runtime) LaunchInstance(packageID string) (Launch, error) {
	pkg, ok := r.packages[packageID]
	if !ok {
		return Launch{}, ErrPackageNotInstalled
	}
	return r.start(pkg, false)
}

func (r *Runtime) start(pkg *AppPackage, primary bool) (Launch, error) {
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return Launch{}, err
	}
	process := r.processes.Register(pkg.ID)
	r.byProcess[process.PID] = map[WindowID]struct{}{}
	if primary {
		r.byPackage[pkg.ID] = process.PID
	}
	r.credentials[process.PID] = secret
	return Launch{Session: Session{pid: process.PID, secret: secret}, Package: pkg, State: append([]byte(nil), r.saved[pkg.ID]...)}, nil
}

// AttachWindow registers a window for the calling application only.
func (r *Runtime) AttachWindow(s Session, id WindowID, displayID DisplayID, config WindowConfig) bool {
	if !r.authorized(s) {
		return false
	}
	display, ok := r.displays[displayID]
	if !ok || !fits(display.config.Bounds(), config.Bounds) {
		return false
	}
	if r.windowExists(id) {
		return false
	}
	display.windows.Register(id, config)
	r.owner[id] = s.pid
	r.displayOf[id] = displayID
	windows := r.byProcess[s.pid]
	windows[id] = struct{}{}
	return true
}

// ResizeWindow changes the bounds of a window owned by the calling app. The
// next submitted buffer must match the new bounds.
func (r *Runtime) ResizeWindow(s Session, id WindowID, bounds geom.Rect) bool {
	if !r.owns(s, id) {
		return false
	}
	display := r.displays[r.displayOf[id]]
	if !fits(display.config.Bounds(), bounds) {
		return false
	}
	return display.windows.Resize(id, bounds)
}

// HideWindow removes one owned window from presentation without terminating
// its process. A later Activate makes it visible again.
func (r *Runtime) HideWindow(s Session, id WindowID) bool {
	if !r.owns(s, id) {
		return false
	}
	if !r.displays[r.displayOf[id]].windows.Hide(id) {
		return false
	}
	r.displays[r.displayOf[id]].input.Release(id)
	if !r.hasVisibleWindow(s.pid) {
		r.processes.Cache(s.pid)
	}
	return true
}

// Split places two existing windows side by side on one display. It is a
// trusted system operation; applications may resize only their own window.
func (r *Runtime) Split(displayID DisplayID, first, second WindowID) bool {
	if first == second || r.displayOf[first] != displayID || r.displayOf[second] != displayID {
		return false
	}
	display, ok := r.displays[displayID]
	if !ok {
		return false
	}
	width := display.config.Width / 2
	left := geom.Rect{Width: width, Height: display.config.Height}
	right := geom.Rect{X: width, Width: display.config.Width - width, Height: display.config.Height}
	if !display.windows.Resize(first, left) || !display.windows.Resize(second, right) {
		return false
	}
	return display.windows.Activate(first) && display.windows.Activate(second)
}

// ResizeSplit moves the divider between two windows already sharing a display.
func (r *Runtime) ResizeSplit(displayID DisplayID, first, second WindowID, divider int) bool {
	display, ok := r.displays[displayID]
	if !ok || first == second || r.displayOf[first] != displayID || r.displayOf[second] != displayID ||
		divider <= 0 || divider >= display.config.Width {
		return false
	}
	left := geom.Rect{Width: divider, Height: display.config.Height}
	right := geom.Rect{X: divider, Width: display.config.Width - divider, Height: display.config.Height}
	return display.windows.Resize(first, left) && display.windows.Resize(second, right)
}

// SwapWindows exchanges the placements of two windows on one display.
func (r *Runtime) SwapWindows(displayID DisplayID, first, second WindowID) bool {
	display, ok := r.displays[displayID]
	if !ok || first == second || r.displayOf[first] != displayID || r.displayOf[second] != displayID {
		return false
	}
	a, b := display.windows.Windows[first], display.windows.Windows[second]
	if a == nil || b == nil {
		return false
	}
	a.Bounds, b.Bounds = b.Bounds, a.Bounds
	a.Buffer, b.Buffer = Buffer{}, Buffer{}
	return true
}

// AddDisplay connects a new logical display with an independent window scene.
func (r *Runtime) AddDisplay(id DisplayID, config DisplayConfig) bool {
	if id == "" || config.Width <= 0 || config.Height <= 0 || config.Revision <= 0 {
		return false
	}
	if _, exists := r.displays[id]; exists {
		return false
	}
	r.displays[id] = &display{config: config, windows: NewWindowServer(), input: NewInputDispatcher()}
	return true
}

// RemoveDisplay disconnects an empty external display. The primary display is
// permanent, and callers must move every window before disconnecting another.
func (r *Runtime) RemoveDisplay(id DisplayID) bool {
	if id == PrimaryDisplay {
		return false
	}
	display, ok := r.displays[id]
	if !ok || len(display.windows.Windows) != 0 {
		return false
	}
	delete(r.displays, id)
	delete(r.recordings, id)
	return true
}

// MoveWindow moves an existing window to another display. This is a trusted
// system operation; the next app buffer must match bounds on the new display.
func (r *Runtime) MoveWindow(id WindowID, to DisplayID, bounds geom.Rect) bool {
	from, ok := r.displayOf[id]
	if !ok || from == to {
		return false
	}
	target, ok := r.displays[to]
	if !ok || !fits(target.config.Bounds(), bounds) {
		return false
	}
	window := r.displays[from].windows.Detach(id)
	if window == nil {
		return false
	}
	r.displays[from].input.Release(id)
	window.Bounds = bounds
	window.Buffer = Buffer{}
	target.windows.Attach(window)
	r.displayOf[id] = to
	return true
}

// Terminate removes one process and every window it owns. It is a trusted
// system lifecycle operation; applications cannot choose another process by
// presenting a Session.
func (r *Runtime) Terminate(pid int) bool {
	process, ok := r.processes.processes[pid]
	if !ok {
		return false
	}
	for id := range r.byProcess[pid] {
		display := r.displays[r.displayOf[id]]
		display.input.Release(id)
		display.windows.Unregister(id)
		delete(r.owner, id)
		delete(r.displayOf, id)
	}
	delete(r.byProcess, pid)
	if r.byPackage[process.PackageID] == pid {
		delete(r.byPackage, process.PackageID)
	}
	delete(r.credentials, pid)
	r.processes.Unregister(pid)
	return true
}

// Evict removes a cached process under system memory pressure. The retained
// record makes the cold-launch boundary observable without retaining app data.
func (r *Runtime) Evict(pid int) bool {
	process, ok := r.processes.processes[pid]
	if !ok || !r.processes.Evict(pid) {
		return false
	}
	for id := range r.byProcess[pid] {
		display := r.displays[r.displayOf[id]]
		display.input.Release(id)
		display.windows.Unregister(id)
		delete(r.owner, id)
		delete(r.displayOf, id)
	}
	delete(r.byProcess, pid)
	delete(r.credentials, pid)
	if r.byPackage[process.PackageID] == pid {
		delete(r.byPackage, process.PackageID)
	}
	return true
}

// Activate foregrounds a window only when the caller owns it.
func (r *Runtime) Activate(s Session, id WindowID) bool {
	if !r.owns(s, id) {
		return false
	}
	display := r.displays[r.displayOf[id]]
	previous := display.windows.Windows[display.windows.Foreground]
	if previous != nil {
		if _, appOwned := r.owner[previous.ID]; !appOwned {
			previous = nil
		}
	}
	if !display.windows.Activate(id) {
		return false
	}
	display.transition.start(previous, id)
	if !r.processes.Activate(s.pid) {
		return false
	}
	return true
}

// Back restores the previously active window and its owning process.
func (r *Runtime) Back() (WindowID, bool) {
	display := r.displays[PrimaryDisplay]
	id, ok := display.windows.Back()
	if !ok {
		return "", false
	}
	pid, ok := r.owner[id]
	if !ok || !r.processes.Activate(pid) {
		return "", false
	}
	return id, true
}

// Submit accepts a buffer only for a window the caller owns.
func (r *Runtime) Submit(s Session, id WindowID, buffer Buffer) bool {
	if !r.owns(s, id) {
		return false
	}
	if !r.acceptsBuffer(s, buffer) {
		return false
	}
	display := r.displays[r.displayOf[id]]
	window := display.windows.Windows[id]
	if buffer.Bounds.Width != window.Bounds.Width ||
		buffer.Bounds.Height != window.Bounds.Height {
		return false
	}
	return display.windows.SubmitBuffer(id, buffer)
}

// Present submits a buffer and composes a frame. It reports success only when
// the resulting frame contains the submitted window revision.
func (r *Runtime) Present(s Session, id WindowID, buffer Buffer) (Frame, bool) {
	if !r.Submit(s, id, buffer) {
		return r.LastFrame(r.displayOf[id]), false
	}
	frame := r.VSync(r.displayOf[id])
	return frame, frame.Presents(id, buffer.Revision)
}

// Input dispatches normalized input to one visible window on a display.
func (r *Runtime) Input(displayID DisplayID, event InputEvent) (WindowInput, bool) {
	display, ok := r.displays[displayID]
	if !ok {
		return WindowInput{}, false
	}
	return display.input.Dispatch(event, &display.windows)
}

func (r *Runtime) authorized(s Session) bool {
	secret, ok := r.credentials[s.pid]
	state, live := r.processes.State(s.pid)
	return ok && live && state != Crashed && subtle.ConstantTimeCompare(secret[:], s.secret[:]) == 1
}

func (r *Runtime) owns(s Session, id WindowID) bool {
	return r.authorized(s) && r.owner[id] == s.pid
}

func (r *Runtime) session(pid int) Session {
	return Session{pid: pid, secret: r.credentials[pid]}
}

func (r *Runtime) FocusedWindowID(displayID DisplayID) (WindowID, bool) {
	display, ok := r.displays[displayID]
	if !ok {
		return "", false
	}
	return display.windows.FocusedWindowID()
}

func (r *Runtime) Compose(displayID DisplayID) Frame {
	return r.VSync(displayID)
}

// VSync adopts the latest submitted buffers into the next display frame.
// The POC calls it synchronously.
func (r *Runtime) VSync(displayID DisplayID) Frame {
	display, ok := r.displays[displayID]
	if !ok {
		return Frame{}
	}
	if display.config.Power == ScreenOff {
		display.compositor.LastFrame = Frame{Number: display.compositor.LastFrame.Number + 1, DisplayID: displayID, Display: display.config}
		return display.compositor.LastFrame
	}
	frame := display.compositor.Compose(displayID, display.windows, display.config)
	frame.Layers = display.transition.apply(frame.Layers, time.Now())
	display.compositor.LastFrame = frame
	r.record(displayID)
	return frame
}

func (r *Runtime) LastFrame(displayID DisplayID) Frame {
	display, ok := r.displays[displayID]
	if !ok {
		return Frame{}
	}
	return display.compositor.LastFrame
}

func (r *Runtime) Foreground(displayID DisplayID) WindowID {
	display, ok := r.displays[displayID]
	if !ok {
		return ""
	}
	return display.windows.Foreground
}

func (r *Runtime) Display(id DisplayID) (DisplayConfig, bool) {
	display, ok := r.displays[id]
	if !ok {
		return DisplayConfig{}, false
	}
	return display.config, true
}

// Config reports the current server-owned system configuration.
func (r *Runtime) Config() SystemConfig { return r.config }

// SetConfig applies the trusted system theme and advances its revision.
func (r *Runtime) SetConfig(config SystemConfig) SystemConfig {
	config.Revision = r.config.Revision + 1
	r.config = config
	return r.config
}

func (r *Runtime) Rotate(id DisplayID) (DisplayConfig, bool) {
	if r.config.OrientationLocked {
		return DisplayConfig{}, false
	}
	display, ok := r.displays[id]
	if !ok {
		return DisplayConfig{}, false
	}
	display.config = display.config.Rotate()
	return display.config, true
}

func (r *Runtime) SetScreenPower(id DisplayID, power ScreenPower) (DisplayConfig, bool) {
	display, ok := r.displays[id]
	if !ok {
		return DisplayConfig{}, false
	}
	display.config.Power = power
	display.config.Revision++
	return display.config, true
}

// RegisterSystemWindow adds a window owned exclusively by trusted system UI.
func (r *Runtime) RegisterSystemWindow(id WindowID, displayID DisplayID, bounds geom.Rect) bool {
	display, ok := r.displays[displayID]
	if !ok || r.windowExists(id) || !fits(display.config.Bounds(), bounds) {
		return false
	}
	display.windows.Register(id, WindowConfig{Bounds: bounds})
	r.displayOf[id] = displayID
	return true
}

// ResizeSystemWindow changes the bounds of a trusted system-owned window.
func (r *Runtime) ResizeSystemWindow(id WindowID, bounds geom.Rect) bool {
	displayID, ok := r.displayOf[id]
	if !ok {
		return false
	}
	if _, appOwned := r.owner[id]; appOwned {
		return false
	}
	display := r.displays[displayID]
	if !fits(display.config.Bounds(), bounds) {
		return false
	}
	return display.windows.Resize(id, bounds)
}

func (r *Runtime) SubmitSystem(id WindowID, buffer Buffer) bool {
	displayID, ok := r.displayOf[id]
	if !ok {
		return false
	}
	if _, appOwned := r.owner[id]; appOwned {
		return false
	}
	display := r.displays[displayID]
	window := display.windows.Windows[id]
	if window == nil || buffer.Bounds.Width != window.Bounds.Width || buffer.Bounds.Height != window.Bounds.Height {
		return false
	}
	// System overlays are regenerated from display state, whose revision may be
	// unchanged when an overlay is reopened. The trusted runtime assigns the
	// next surface revision rather than weakening stale-buffer checks for apps.
	if buffer.Revision <= window.Buffer.Revision {
		buffer.Revision = window.Buffer.Revision + 1
	}
	return display.windows.SubmitBuffer(id, buffer)
}

func (r *Runtime) ShowSystemWindow(id WindowID) bool {
	displayID, ok := r.displayOf[id]
	if !ok {
		return false
	}
	if _, appOwned := r.owner[id]; appOwned {
		return false
	}
	return r.displays[displayID].windows.Activate(id)
}

func (r *Runtime) HideSystemWindow(id WindowID) bool {
	displayID, ok := r.displayOf[id]
	if !ok {
		return false
	}
	if _, appOwned := r.owner[id]; appOwned {
		return false
	}
	windows := &r.displays[displayID].windows
	if !windows.Hide(id) {
		return false
	}
	r.displays[displayID].input.Release(id)
	windows.Forget(id)
	return true
}

func (r *Runtime) ProtectWindow(id WindowID) bool {
	if _, ok := r.displayOf[id]; !ok {
		return false
	}
	if _, appOwned := r.owner[id]; appOwned {
		return false
	}
	r.protected[id] = true
	return true
}

// SetContentProtected lets an app protect its own window from screenshots and
// recordings. A session cannot mark another app's window protected.
func (r *Runtime) SetContentProtected(s Session, id WindowID, protected bool) bool {
	if !r.owns(s, id) {
		return false
	}
	if protected {
		r.protected[id] = true
	} else {
		delete(r.protected, id)
	}
	return true
}

// Capture returns the latest composed frame with protected windows omitted.
func (r *Runtime) Capture(displayID DisplayID) (Frame, bool) {
	display, ok := r.displays[displayID]
	if !ok || display.compositor.LastFrame.DisplayID != displayID {
		return Frame{}, false
	}
	return display.compositor.Capture(r.protected), true
}

func (r *Runtime) windowExists(id WindowID) bool {
	_, ok := r.displayOf[id]
	return ok
}

func (r *Runtime) hasVisibleWindow(pid int) bool {
	for id := range r.byProcess[pid] {
		display := r.displays[r.displayOf[id]]
		if window := display.windows.Windows[id]; window != nil && window.Visible {
			return true
		}
	}
	return false
}

func (r *Runtime) ProcessState(pid int) (ProcessState, bool) {
	return r.processes.State(pid)
}

func (r *Runtime) MarkUnresponsive(pid int) bool { return r.processes.SetState(pid, Unresponsive) }
func (r *Runtime) Crash(pid int) bool            { return r.processes.SetState(pid, Crashed) }

// SaveState retains app-provided state only for a later cold relaunch. The
// server treats it as opaque bytes and never returns it through its APIs.
func (r *Runtime) SaveState(s Session, state []byte) bool {
	if !r.authorized(s) {
		return false
	}
	process, ok := r.processes.processes[s.pid]
	if !ok {
		return false
	}
	r.saved[process.PackageID] = append([]byte(nil), state...)
	return true
}

func fits(outer, inner geom.Rect) bool {
	return inner.Width > 0 && inner.Height > 0 && inner.X >= outer.X && inner.Y >= outer.Y &&
		inner.X+inner.Width <= outer.X+outer.Width && inner.Y+inner.Height <= outer.Y+outer.Height
}
