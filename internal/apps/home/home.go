// Package home models the system home-screen application.
package home

import (
	"mmos/internal/client"
	"mmos/internal/common/geom"
	"mmos/internal/common/model"
	"mmos/internal/common/protocol"
)

const (
	OpenChat     client.Command = "open_chat"
	OpenSettings client.Command = "open_settings"
)

// Launcher is Home's view of the system launch endpoint.
type Launcher interface {
	Launch(protocol.LaunchRequest) protocol.LaunchReply
}

// App is the Home application's client-side instance.
type App struct {
	Process  *client.AppProcess
	launcher Launcher
}

// New creates the Home application. launcher is the process's connection to
// the system launch endpoint.
func New(pkg *model.AppPackage, launcher Launcher) *App {
	window := client.NewWindow("home", "Home")
	window.Root.Children = []*client.UINode{
		{
			ID: "chat", Kind: client.Button,
			Bounds: geom.Rect{X: 100, Y: 180, Width: 120, Height: 60},
			Text:   "Chat", Visible: true, Enabled: true, Command: OpenChat,
		},
		{
			ID: "settings", Kind: client.Button,
			Bounds: geom.Rect{X: 100, Y: 260, Width: 120, Height: 60},
			Text:   "Settings", Visible: true, Enabled: true, Command: OpenSettings,
		},
	}
	return &App{
		Process: &client.AppProcess{
			Package: pkg,
			Windows: map[model.WindowID]*client.Window{window.ID: window},
		},
		launcher: launcher,
	}
}

// HandleInput interprets commands from the Home window and starts the
// selected installed app through the system-provided launch function.
func (a *App) HandleInput(windowID model.WindowID, event model.InputEvent) bool {
	window, ok := a.Process.Windows[windowID]
	if !ok {
		return false
	}
	command, ok := window.CommandAt(event)
	if !ok {
		return false
	}
	switch command {
	case OpenChat:
		return a.launch("chat-app")
	case OpenSettings:
		return a.launch("settings-app")
	}
	return false
}

func (a *App) launch(packageID string) bool {
	if a.launcher == nil {
		return false
	}
	return a.launcher.Launch(protocol.LaunchRequest{PackageID: packageID}).Started
}
