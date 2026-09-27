// Package model defines the contract shared by clients and the server.
package model

import "mmos/internal/common/geom"

type AppPackage struct{ ID, EntryPoint string }
type Theme uint8

const (
	LightTheme Theme = iota
	DarkTheme
)

type SystemConfig struct {
	Revision uint64
	Theme    Theme
}

func DefaultSystemConfig() SystemConfig { return SystemConfig{Revision: 1, Theme: LightTheme} }

type WindowID string
type Buffer struct {
	Revision int
	Content  string
	Bounds   geom.Rect
}
type InputEvent interface{ isInputEvent() }
type PointerAction uint8

const (
	PointerDown PointerAction = iota
	PointerMove
	PointerUp
	PointerCancel
)

type PointerSource uint8

const (
	TouchSource PointerSource = iota
	StylusSource
	MouseSource
	TrackpadSource
)

type PointerSample struct {
	ID       int
	X, Y     int
	Pressure float32
}
type PointerEvent struct {
	Action           PointerAction
	Source           PointerSource
	ChangedPointerID int
	Pointers         []PointerSample
}

func (PointerEvent) isInputEvent() {}
func (e PointerEvent) PointerByID(id int) (PointerSample, bool) {
	for _, p := range e.Pointers {
		if p.ID == id {
			return p, true
		}
	}
	return PointerSample{}, false
}

type KeyAction uint8

const (
	KeyDown KeyAction = iota
	KeyUp
)

type KeyEvent struct {
	Action      KeyAction
	Key         string
	RepeatCount int
}

func (KeyEvent) isInputEvent() {}

type TextEvent struct {
	Text      string
	Composing bool
}

func (TextEvent) isInputEvent() {}

type ScrollEvent struct {
	Source         PointerSource
	X, Y           int
	DeltaX, DeltaY int
}

func (ScrollEvent) isInputEvent() {}
