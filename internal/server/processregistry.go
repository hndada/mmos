package server

// Session is an opaque server-issued capability for one application process.
// Applications can present it to the server but cannot inspect or forge it.
type Session struct {
	pid    int
	secret [32]byte
}

func (s Session) PID() int { return s.pid }

type ProcessState uint8

const (
	// Starting has a process but no visible window yet.
	Starting ProcessState = iota
	Active
	// Cached retains a process with no visible window so a later launch can resume it.
	Cached
	// Evicted retains only the lifecycle record; its process and windows are gone.
	Evicted
	Unresponsive
	Crashed
)

// Process is the server-side record for an application execution.
// Its PID and lifecycle state are server-owned.
type Process struct {
	PID       int
	PackageID string
	State     ProcessState
}

// ProcessRegistry owns server-side process records and PID allocation.
type ProcessRegistry struct {
	processes map[int]Process
	nextPID   int
}

func NewProcessRegistry() ProcessRegistry {
	return ProcessRegistry{
		processes: map[int]Process{},
		nextPID:   1,
	}
}

// Register allocates a PID and retains the record for a started package.
func (r *ProcessRegistry) Register(packageID string) Process {
	process := Process{PID: r.nextPID, PackageID: packageID, State: Starting}
	r.nextPID++
	r.processes[process.PID] = process
	return process
}

func (r *ProcessRegistry) Evict(pid int) bool {
	process, ok := r.processes[pid]
	if !ok || process.State != Cached {
		return false
	}
	process.State = Evicted
	r.processes[pid] = process
	return true
}

func (r *ProcessRegistry) Unregister(pid int) {
	delete(r.processes, pid)
}

// Activate marks a live process as having a visible window. More than one
// process may be active while their windows share a display.
func (r *ProcessRegistry) Activate(pid int) bool {
	process, ok := r.processes[pid]
	if !ok || process.State == Evicted || process.State == Crashed {
		return false
	}
	process.State = Active
	r.processes[pid] = process
	return true
}

// Cache marks a live process with no visible windows as resumable.
func (r *ProcessRegistry) Cache(pid int) bool {
	process, ok := r.processes[pid]
	if !ok || process.State == Evicted || process.State == Crashed {
		return false
	}
	process.State = Cached
	r.processes[pid] = process
	return true
}

func (r *ProcessRegistry) SetState(pid int, state ProcessState) bool {
	process, ok := r.processes[pid]
	if !ok {
		return false
	}
	process.State = state
	r.processes[pid] = process
	return true
}

func (r *ProcessRegistry) Has(pid int) bool {
	_, ok := r.processes[pid]
	return ok
}

func (r *ProcessRegistry) State(pid int) (ProcessState, bool) {
	process, ok := r.processes[pid]
	return process.State, ok
}

func (r *ProcessRegistry) Len() int {
	return len(r.processes)
}
