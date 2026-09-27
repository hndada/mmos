// Package notice implements status and notification surfaces.
package notice

import (
	"errors"
	"fmt"

	"mmos/internal/common/geom"
	"mmos/internal/server"
)

var ErrInstall = errors.New("install notice")

const (
	status server.WindowID = "status"
	window server.WindowID = "notices"
)

type App struct {
	runtime *server.Runtime
	display server.DisplayID
	open    bool
	pull    struct {
		active bool
		x, y   int
	}
}

func New(runtime *server.Runtime, display server.DisplayID) (*App, error) {
	if runtime == nil {
		return nil, ErrInstall
	}
	config, ok := runtime.Display(display)
	bounds := geom.Rect{Width: config.Width, Height: 24}
	if !ok || !runtime.RegisterSystemWindow(status, display, bounds) || !runtime.RegisterSystemWindow(window, display, config.Bounds()) ||
		!runtime.SubmitSystem(status, server.Buffer{Revision: 1, Content: "status", Bounds: bounds}) || !runtime.ShowSystemWindow(status) || !runtime.HideSystemWindow(window) {
		return nil, ErrInstall
	}
	return &App{runtime: runtime, display: display}, nil
}

func (a *App) HandleSystem(event server.SystemEvent) (server.Frame, bool) {
	if event.Action == server.SystemBack && a.open {
		return a.hide(), true
	}
	if event.Action == server.SystemNotices {
		if a.open {
			return a.hide(), true
		}
		return a.show()
	}
	return a.runtime.LastFrame(a.display), false
}

func (a *App) HandleInput(event server.InputEvent) (server.Frame, bool) {
	if a.open {
		return a.runtime.LastFrame(a.display), true
	}
	_, ok := a.runtime.Display(a.display)
	if !ok {
		return a.runtime.LastFrame(a.display), false
	}
	p, ok := event.(server.PointerEvent)
	if !ok || p.ChangedPointerID != 0 {
		return a.runtime.LastFrame(a.display), false
	}
	s, ok := p.PointerByID(0)
	if !ok {
		return a.runtime.LastFrame(a.display), false
	}
	switch p.Action {
	case server.PointerDown:
		a.pull.active = s.Y < 24
		a.pull.x, a.pull.y = s.X, s.Y
	case server.PointerCancel:
		a.pull.active = false
	case server.PointerUp:
		active := a.pull.active
		a.pull.active = false
		if active && s.Y-a.pull.y >= 96 && abs(s.X-a.pull.x) <= 80 {
			return a.show()
		}
	}
	return a.runtime.LastFrame(a.display), false
}

func (a *App) Resize() bool {
	config, ok := a.runtime.Display(a.display)
	if !ok || !a.runtime.ResizeSystemWindow(status, geom.Rect{Width: config.Width, Height: 24}) || !a.runtime.ResizeSystemWindow(window, config.Bounds()) {
		return false
	}
	return a.Refresh()
}

func (a *App) Refresh() bool {
	config, ok := a.runtime.Display(a.display)
	if !ok {
		return false
	}
	content := "status"
	if recording, ok := a.runtime.Recording(a.display); ok && recording.Active {
		content = "status(recording)"
	}
	return a.runtime.SubmitSystem(status, server.Buffer{Revision: config.Revision, Content: content, Bounds: geom.Rect{Width: config.Width, Height: 24}})
}

func (a *App) show() (server.Frame, bool) {
	config, ok := a.runtime.Display(a.display)
	if !ok || !a.runtime.SubmitSystem(window, server.Buffer{Revision: config.Revision, Content: a.content(), Bounds: config.Bounds()}) || !a.runtime.ShowSystemWindow(window) {
		return a.runtime.LastFrame(a.display), true
	}
	a.open = true
	return a.runtime.VSync(a.display), true
}

func (a *App) hide() server.Frame {
	a.open = false
	if !a.runtime.HideSystemWindow(window) {
		return a.runtime.LastFrame(a.display)
	}
	return a.runtime.VSync(a.display)
}
func (a *App) content() string {
	notices := a.runtime.Notices()
	if len(notices) == 0 {
		return "notices(count=0)"
	}
	latest := notices[len(notices)-1]
	return fmt.Sprintf("notices(count=%d, latest=%s: %s)", len(notices), latest.Title, latest.Body)
}
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
