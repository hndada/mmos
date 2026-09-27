// Package chat models the Chat application built on the generic client runtime.
package chat

import (
	"fmt"

	"mmos/internal/client"
	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

const (
	SendMessage client.Command = "send_message"
)

type State struct {
	MessageCount int
}

type App struct {
	Process *client.AppProcess
	State   *State
}

func New(pkg *model.AppPackage) *App {
	window := newWindow()
	return &App{
		Process: &client.AppProcess{
			Package: pkg,
			Windows: map[model.WindowID]*client.Window{window.ID: window},
		},
		State: &State{},
	}
}

func (a *App) HandleInput(windowID model.WindowID, event model.InputEvent) bool {
	window, ok := a.Process.Windows[windowID]
	if !ok {
		return false
	}
	command, ok := window.CommandAt(event)
	if !ok {
		return false
	}
	if command != SendMessage {
		return false
	}
	a.State.MessageCount++
	window.Node("message").Text = fmt.Sprintf("Messages sent: %d", a.State.MessageCount)
	return true
}

func newWindow() *client.Window {
	window := client.NewWindow("chat", "Chat")
	window.Root.Children = []*client.UINode{
		{ID: "message", Kind: client.Text, Bounds: geom.Rect{X: 20, Y: 70, Width: 280, Height: 40}, Text: "No messages sent.", Visible: true},
		{
			ID:      "send",
			Kind:    client.Button,
			Bounds:  geom.Rect{X: 100, Y: 200, Width: 120, Height: 60},
			Text:    "Send",
			Visible: true,
			Enabled: true,
			Command: SendMessage,
		},
	}
	return window
}
