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
	MessageCount  int
	Draft         string
	Composition   string
	Editing       bool
	KeyboardInset int
	pages         []string
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
	if text, ok := event.(model.TextEvent); ok {
		return a.handleText(window, text)
	}
	if key, ok := event.(model.KeyEvent); ok {
		return a.handleKey(window, key)
	}
	if a.focusesEditor(window, event) {
		a.State.Editing = true
		a.showDraft(window)
		return true
	}
	command, ok := window.CommandAt(event)
	if !ok {
		return false
	}
	switch command {
	case SendMessage:
		a.send(window)
		return true
	case OpenDetails:
		a.State.pages = append(a.State.pages, "details")
		window.Node("message").Text = "Chat details"
		return true
	}
	return false
}

func (a *App) Editing() bool { return a.State.Editing }

// SetKeyboardInset moves the composer into the remaining visible area. The
// app keeps its window bounds; the inset is owned by the system IME overlay.
func (a *App) SetKeyboardInset(inset int) {
	if inset < 0 {
		inset = 0
	}
	a.State.KeyboardInset = inset
	window := a.Process.Windows["chat"]
	if window == nil {
		return
	}
	editor, send := window.Node("editor"), window.Node("send")
	if editor == nil || send == nil {
		return
	}
	if inset == 0 {
		editor.Bounds.Y = 130
		send.Bounds.Y = 200
		return
	}
	bottom := window.Bounds().Height - inset - 16
	send.Bounds.Y = bottom - send.Bounds.Height
	editor.Bounds.Y = send.Bounds.Y - editor.Bounds.Height - 8
}

func (a *App) handleText(window *client.Window, event model.TextEvent) bool {
	if !a.State.Editing {
		return false
	}
	if event.Composing {
		a.State.Composition = event.Text
	} else {
		a.State.Draft += event.Text
		a.State.Composition = ""
	}
	a.showDraft(window)
	return true
}

func (a *App) handleKey(window *client.Window, event model.KeyEvent) bool {
	if !a.State.Editing || event.Action != model.KeyDown {
		return false
	}
	switch event.Key {
	case "Backspace":
		if a.State.Composition != "" {
			a.State.Composition = ""
		} else {
			a.State.Draft = dropLastRune(a.State.Draft)
		}
	case "Enter":
		a.send(window)
	default:
		return false
	}
	a.showDraft(window)
	return true
}

func (a *App) focusesEditor(window *client.Window, event model.InputEvent) bool {
	p, ok := event.(model.PointerEvent)
	if !ok || p.Action != model.PointerUp || p.ChangedPointerID != 0 {
		return false
	}
	sample, ok := p.PointerByID(0)
	return ok && window.Node("editor").Bounds.Contains(sample.X, sample.Y)
}

func (a *App) send(window *client.Window) {
	if a.State.Composition != "" {
		a.State.Draft += a.State.Composition
		a.State.Composition = ""
	}
	a.State.MessageCount++
	a.State.Draft = ""
	window.Node("message").Text = fmt.Sprintf("Messages sent: %d", a.State.MessageCount)
	a.showDraft(window)
}

func (a *App) showDraft(window *client.Window) {
	text := a.State.Draft + a.State.Composition
	if text == "" {
		text = "Type a message"
	}
	window.Node("editor").Text = text
}

func dropLastRune(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	return string(runes[:len(runes)-1])
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
		{ID: "editor", Kind: client.Text, Bounds: geom.Rect{X: 20, Y: 130, Width: 280, Height: 48}, Text: "Type a message", Visible: true},
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
