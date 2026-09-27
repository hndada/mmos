// Package history implements the Recents task-card surface.
package history

import (
	"errors"
	"fmt"

	"mmos/internal/server"
)

var ErrInstall = errors.New("install history")

const window server.WindowID = "history"

type Entry struct {
	WindowID server.WindowID
	Label    string
}

// App owns the Recents cards and returns the selected task to composition.
type App struct {
	runtime *server.Runtime
	display server.DisplayID
	entries []Entry
	open    bool
}

func New(runtime *server.Runtime, display server.DisplayID) (*App, error) {
	if runtime == nil {
		return nil, ErrInstall
	}
	config, ok := runtime.Display(display)
	if !ok || !runtime.RegisterSystemWindow(window, display, config.Bounds()) || !runtime.HideSystemWindow(window) {
		return nil, ErrInstall
	}
	return &App{runtime: runtime, display: display}, nil
}

func (a *App) Show(entries []Entry) (server.Frame, bool) {
	a.entries = append(a.entries[:0], entries...)
	config, ok := a.runtime.Display(a.display)
	if !ok || !a.runtime.SubmitSystem(window, server.Buffer{Revision: config.Revision, Content: a.content(), Bounds: config.Bounds()}) || !a.runtime.ShowSystemWindow(window) {
		return a.runtime.LastFrame(a.display), false
	}
	a.open = true
	return a.runtime.VSync(a.display), true
}

// HandleInput consumes visible-card input and returns the selected task ID.
func (a *App) HandleInput(event server.InputEvent) (server.WindowID, server.Frame, bool) {
	if !a.open {
		return "", a.runtime.LastFrame(a.display), false
	}
	p, ok := event.(server.PointerEvent)
	if !ok || p.Action != server.PointerUp || p.ChangedPointerID != 0 {
		return "", a.runtime.LastFrame(a.display), true
	}
	sample, ok := p.PointerByID(0)
	if !ok {
		return "", a.runtime.LastFrame(a.display), true
	}
	index := (sample.Y - 80) / 96
	if index < 0 || index >= len(a.entries) {
		return "", a.runtime.LastFrame(a.display), true
	}
	id := a.entries[index].WindowID
	a.open = false
	if !a.runtime.HideSystemWindow(window) {
		return "", a.runtime.LastFrame(a.display), true
	}
	return id, a.runtime.VSync(a.display), true
}

func (a *App) Resize() bool {
	config, ok := a.runtime.Display(a.display)
	if !ok || !a.runtime.ResizeSystemWindow(window, config.Bounds()) {
		return false
	}
	if !a.open {
		return true
	}
	_, ok = a.Show(a.entries)
	return ok
}

func (a *App) content() string {
	labels := make([]string, len(a.entries))
	for i, entry := range a.entries {
		labels[i] = entry.Label
	}
	return fmt.Sprintf("recents(count=%d, apps=%v)", len(labels), labels)
}
