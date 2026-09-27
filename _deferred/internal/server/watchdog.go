package server

import "time"

// Watchdog tracks foreground response deadlines. It has no goroutine: the
// trusted event loop supplies time through Check, making shutdown and tests
// deterministic.
type Watchdog struct {
	budget   time.Duration
	deadline map[int]time.Time
}

func NewWatchdog(budget time.Duration) Watchdog {
	return Watchdog{budget: budget, deadline: map[int]time.Time{}}
}

// Observe starts or renews a process response deadline.
func (w *Watchdog) Observe(pid int, now time.Time) {
	if w.budget <= 0 {
		return
	}
	w.deadline[pid] = now.Add(w.budget)
}

// Check returns each process whose deadline elapsed and forgets it. The
// runtime remains responsible for changing process state and showing UI.
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

// CheckWatchdog marks active processes unresponsive after the supplied
// watchdog has detected missed work. A caller can then use Crash explicitly.
func (r *Runtime) CheckWatchdog(w *Watchdog, now time.Time) []int {
	if w == nil {
		return nil
	}
	var unresponsive []int
	for _, pid := range w.Check(now) {
		if r.MarkUnresponsive(pid) {
			unresponsive = append(unresponsive, pid)
		}
	}
	return unresponsive
}
