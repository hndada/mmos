package server

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
)

var ErrPackageNotInstalled = errors.New("package is not installed")
var ErrDisplayNotFound = errors.New("display not found")

// Launch is the result of starting or resuming an installed package.
type Launch struct {
	Session Session
	Package *AppPackage
	Reused  bool
}

// Runtime owns the server-side process and display lifecycle for one system.
// Client applications create their own UI, then attach submitted windows to
// the process that the runtime started.
type Runtime struct {
	packages    map[string]*AppPackage
	config      SystemConfig
	processes   ProcessRegistry
	input       InputServer
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
	compositor Compositor
}

func NewRuntime(installed ...*AppPackage) Runtime {
	runtime := Runtime{
		packages:  make(map[string]*AppPackage, len(installed)),
		config:    DefaultSystemConfig(),
		processes: NewProcessRegistry(),
		input:     NewInputServer(),
		displays: map[DisplayID]*display{
			PrimaryDisplay: {
				config:  DisplayConfig{Width: 320, Height: 480, Revision: 1, Density: 1},
				windows: NewWindowServer(),
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
			return Launch{Session: r.session(pid), Package: pkg, Reused: true}, nil
		}
		delete(r.byPackage, packageID)
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return Launch{}, err
	}
	process := r.processes.Register(pkg.ID)
	r.byProcess[process.PID] = map[WindowID]struct{}{}
	r.byPackage[pkg.ID] = process.PID
	r.credentials[process.PID] = secret
	return Launch{
		Session: Session{pid: process.PID, secret: secret},
		Package: pkg,
	}, nil
}

// AttachWindow registers a window for the calling application only.
func (r *Runtime) AttachWindow(s Session, id WindowID, displayID DisplayID, config WindowConfig) bool {
	if !r.authorized(s) {
		return false
	}
	display, ok := r.displays[displayID]
	if !ok || config.Bounds != display.config.Bounds() {
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

// Terminate removes one process and every window it owns. It is a trusted
// system lifecycle operation; applications cannot choose another process by
// presenting a Session.
func (r *Runtime) Terminate(pid int) bool {
	process, ok := r.processes.processes[pid]
	if !ok {
		return false
	}
	for id := range r.byProcess[pid] {
		r.displays[r.displayOf[id]].windows.Unregister(id)
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
		r.displays[r.displayOf[id]].windows.Unregister(id)
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
	if !display.windows.Activate(id) {
		return false
	}
	return r.processes.Activate(s.pid)
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

// Input selects the foreground app window and returns window-local input.
func (r *Runtime) Input(displayID DisplayID, event InputEvent) (WindowInput, bool) {
	display, ok := r.displays[displayID]
	if !ok {
		return WindowInput{}, false
	}
	return r.input.Route(event, &display.windows)
}

func (r *Runtime) authorized(s Session) bool {
	secret, ok := r.credentials[s.pid]
	return ok && subtle.ConstantTimeCompare(secret[:], s.secret[:]) == 1
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
	return display.compositor.Compose(displayID, display.windows, display.config)
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

func (r *Runtime) windowExists(id WindowID) bool {
	_, ok := r.displayOf[id]
	return ok
}

func (r *Runtime) ProcessState(pid int) (ProcessState, bool) {
	return r.processes.State(pid)
}
