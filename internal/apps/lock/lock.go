// Package lock implements the trusted lock and launch-placeholder surfaces.
package lock

import (
	"errors"
	"fmt"

	"mmos/internal/common/geom"
	"mmos/internal/server"
)

var ErrInstall = errors.New("install lock")

const (
	window server.WindowID = "lock"
	splash server.WindowID = "splash"
)

type App struct {
	runtime *server.Runtime
	display server.DisplayID
	locked  bool
}

func New(runtime *server.Runtime, display server.DisplayID) (*App, error) {
	if runtime == nil {
		return nil, ErrInstall
	}
	config, ok := runtime.Display(display)
	if !ok || !runtime.RegisterSystemWindow(window, display, config.Bounds()) ||
		!runtime.RegisterSystemWindow(splash, display, config.Bounds()) || !runtime.ProtectWindow(window) ||
		!runtime.HideSystemWindow(window) || !runtime.HideSystemWindow(splash) {
		return nil, ErrInstall
	}
	return &App{runtime: runtime, display: display}, nil
}

func (a *App) ShowSplash(packageID string) (server.Frame, bool) {
	if !a.show(splash, fmt.Sprintf("splash(package=%s)", packageID)) {
		return a.runtime.LastFrame(a.display), false
	}
	frame := a.runtime.VSync(a.display)
	for _, layer := range frame.Layers {
		if layer.WindowID == splash {
			return frame, true
		}
	}
	return frame, false
}

func (a *App) HideSplash() bool { return a.runtime.HideSystemWindow(splash) }

func (a *App) HandleSystem(event server.SystemEvent) (server.Frame, bool) {
	if a.locked && event.Action != server.SystemUnlock && event.Action != server.SystemScreenOn && event.Action != server.SystemScreenOff {
		return a.runtime.LastFrame(a.display), true
	}
	switch event.Action {
	case server.SystemUnlock:
		return a.unlock()
	case server.SystemScreenOff:
		if a.runtime.Config().LockOnScreenOff && !a.locked && a.show(window, "lock(unlock-button)") {
			a.locked = true
		}
		_, changed := a.runtime.SetScreenPower(a.display, server.ScreenOff)
		return a.runtime.VSync(a.display), changed
	case server.SystemScreenOn:
		_, changed := a.runtime.SetScreenPower(a.display, server.ScreenOn)
		return a.runtime.VSync(a.display), changed
	}
	return a.runtime.LastFrame(a.display), false
}

func (a *App) HandleInput(event server.InputEvent) (server.Frame, bool) {
	if !a.locked {
		return a.runtime.LastFrame(a.display), false
	}
	config, ok := a.runtime.Display(a.display)
	if !ok || !unlockPressed(event, config.Bounds()) {
		return a.runtime.LastFrame(a.display), false
	}
	return a.unlock()
}

func (a *App) Resize() bool {
	config, ok := a.runtime.Display(a.display)
	if !ok || !a.runtime.ResizeSystemWindow(window, config.Bounds()) || !a.runtime.ResizeSystemWindow(splash, config.Bounds()) {
		return false
	}
	return !a.locked || a.show(window, "lock(unlock-button)")
}

func (a *App) show(id server.WindowID, content string) bool {
	config, ok := a.runtime.Display(a.display)
	return ok && a.runtime.SubmitSystem(id, server.Buffer{Revision: config.Revision, Content: content, Bounds: config.Bounds()}) && a.runtime.ShowSystemWindow(id)
}

func (a *App) unlock() (server.Frame, bool) {
	config, ok := a.runtime.Display(a.display)
	if !ok || config.Power == server.ScreenOff || !a.runtime.HideSystemWindow(window) {
		return a.runtime.LastFrame(a.display), false
	}
	a.locked = false
	return a.runtime.VSync(a.display), true
}

func unlockPressed(event server.InputEvent, bounds geom.Rect) bool {
	pointer, ok := event.(server.PointerEvent)
	if !ok || pointer.Action != server.PointerUp || pointer.ChangedPointerID != 0 {
		return false
	}
	sample, ok := pointer.PointerByID(0)
	return ok && geom.Rect{X: (bounds.Width - 240) / 2, Y: bounds.Height - 72, Width: 240, Height: 48}.Contains(sample.X, sample.Y)
}
