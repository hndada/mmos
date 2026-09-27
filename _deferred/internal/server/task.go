package server

import "sort"

// TaskSnapshot is a system-owned summary for History. It deliberately omits
// app memory and Session credentials.
type TaskSnapshot struct {
	PID       int
	PackageID string
	State     ProcessState
}

// Tasks returns stable, PID-ordered snapshots for trusted task UI.
func (r *Runtime) Tasks() []TaskSnapshot {
	tasks := make([]TaskSnapshot, 0, len(r.processes.processes))
	for _, process := range r.processes.processes {
		tasks = append(tasks, TaskSnapshot{PID: process.PID, PackageID: process.PackageID, State: process.State})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].PID < tasks[j].PID })
	return tasks
}
