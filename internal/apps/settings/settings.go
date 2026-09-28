// Package settings models a small client-side settings application.
package settings

import (
	"mmos/internal/client"
	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

const (
	ToggleTheme           client.Command = "toggle_theme"
	ToggleLockOnScreenOff client.Command = "toggle_lock_on_screen_off"
)

// App renders settings controls and asks its composition root to apply a
// changed configuration. The server remains the configuration authority.
type App struct {
	Process *client.AppProcess
	apply   func(model.SystemConfig) bool
}

func New(pkg *model.AppPackage, apply func(model.SystemConfig) bool) *App {
	window := client.NewWindow("settings", "Settings")
	window.Root.Children = []*client.UINode{
		{ID: "title", Kind: client.Text, Bounds: geom.Rect{X: 20, Y: 30, Width: 280, Height: 32}, Text: "Settings", Visible: true},
		{ID: "theme", Kind: client.Button, Bounds: geom.Rect{X: 20, Y: 90, Width: 280, Height: 60}, Visible: true, Enabled: true, Command: ToggleTheme},
		{ID: "lock", Kind: client.Button, Bounds: geom.Rect{X: 20, Y: 170, Width: 280, Height: 60}, Visible: true, Enabled: true, Command: ToggleLockOnScreenOff},
	}
	app := &App{
		Process: &client.AppProcess{Package: pkg, Windows: map[model.WindowID]*client.Window{window.ID: window}},
		apply:   apply,
	}
	app.show()
	return app
}

func (a *App) HandleInput(windowID model.WindowID, event model.InputEvent) bool {
	window, ok := a.Process.Windows[windowID]
	if !ok || a.apply == nil {
		return false
	}
	command, ok := window.CommandAt(event)
	if !ok {
		return false
	}
	config := a.Process.Config
	switch command {
	case ToggleTheme:
		if config.Theme == model.LightTheme {
			config.Theme = model.DarkTheme
		} else {
			config.Theme = model.LightTheme
		}
	case ToggleLockOnScreenOff:
		config.LockOnScreenOff = !config.LockOnScreenOff
	default:
		return false
	}
	return a.apply(config)
}

func (a *App) show() {
	window := a.Process.Windows["settings"]
	if a.Process.Config.Theme == model.DarkTheme {
		window.Node("theme").Text = "Theme: Dark"
	} else {
		window.Node("theme").Text = "Theme: Light"
	}
	if a.Process.Config.LockOnScreenOff {
		window.Node("lock").Text = "Lock after screen off: On"
	} else {
		window.Node("lock").Text = "Lock after screen off: Off"
	}
}

// Configure refreshes the labels from the server-owned configuration.
func (a *App) Configure(config model.SystemConfig) {
	a.Process.Configure(config)
	a.show()
}
