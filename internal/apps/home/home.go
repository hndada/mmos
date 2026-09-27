// Package home models the system home-screen application.
package home

import (
	"mmos/internal/client"
	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

const OpenChat client.Command = "open_chat"

// App is the Home application's client-side instance.
type App struct {
	Process *client.AppProcess
	launch  func() bool
}

// New creates the Home application. launch starts the selected application
// after Home has interpreted an icon command.
func New(pkg *model.AppPackage, launch func() bool) *App {
	window := client.NewWindow("home", "Home")
	window.Root.Children = []*client.UINode{
		{
			ID: "chat", Kind: client.Button,
			Bounds: geom.Rect{X: 100, Y: 180, Width: 120, Height: 60},
			Text:   "Chat", Visible: true, Enabled: true, Command: OpenChat,
		},
	}
	return &App{
		Process: &client.AppProcess{
			Package: pkg,
			Windows: map[model.WindowID]*client.Window{window.ID: window},
		},
		launch: launch,
	}
}

// HandleInput interprets commands from the Home window and starts the
// selected installed app through the system-provided launch function.
func (a *App) HandleInput(windowID model.WindowID, event model.InputEvent) bool {
	window, ok := a.Process.Windows[windowID]
	if !ok || a.launch == nil {
		return false
	}
	command, ok := window.CommandAt(event)
	if !ok || command != OpenChat {
		return false
	}
	return a.launch()
}
