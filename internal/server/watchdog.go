package server

import "time"

type Watchdog struct {
	budget   time.Duration
	deadline map[int]time.Time
}

func NewWatchdog(budget time.Duration) Watchdog {
	return Watchdog{budget: budget, deadline: map[int]time.Time{}}
}
func (w *Watchdog) Observe(pid int, now time.Time) {
	if w.budget > 0 {
		w.deadline[pid] = now.Add(w.budget)
	}
}
func (w *Watchdog) Check(now time.Time) []int {
	var expired []int
	for pid, deadline := range w.deadline {
		if !now.Before(deadline) {
			expired = append(expired, pid)
			delete(w.deadline, pid)
		}
	}
	return expired
}
func (w *Watchdog) Forget(pid int) { delete(w.deadline, pid) }
func (r *Runtime) CheckWatchdog(w *Watchdog, now time.Time) []int {
	if w == nil {
		return nil
	}
	var expired []int
	for _, pid := range w.Check(now) {
		if r.MarkUnresponsive(pid) {
			expired = append(expired, pid)
		}
	}
	return expired
}
