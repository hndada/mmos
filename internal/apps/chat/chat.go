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
	OpenDetails client.Command = "open_details"
)

type State struct {
	MessageCount int
	pages        []string
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
		State: &State{pages: []string{"chat"}},
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
	switch command {
	case SendMessage:
		a.State.MessageCount++
		window.Node("message").Text = fmt.Sprintf("Messages sent: %d", a.State.MessageCount)
		return true
	case OpenDetails:
		a.State.pages = append(a.State.pages, "details")
		window.Node("message").Text = "Chat details"
		return true
	}
	return false
}

func (a *App) Page() string { return a.State.pages[len(a.State.pages)-1] }

// Back consumes app-local history before a system Back action changes apps.
func (a *App) Back() bool {
	if len(a.State.pages) <= 1 {
		return false
	}
	a.State.pages = a.State.pages[:len(a.State.pages)-1]
	a.Process.Windows["chat"].Node("message").Text = "No messages sent."
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
		{
			ID: "details", Kind: client.Button,
			Bounds: geom.Rect{X: 100, Y: 280, Width: 120, Height: 60},
			Text:   "Details", Visible: true, Enabled: true,
			Role: client.ButtonRole, Label: "Open chat details", Command: OpenDetails,
		},
	}
	return window
}
