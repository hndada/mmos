package server

import (
	"mmos/internal/common/geom"
	"mmos/internal/common/model"
)

type (
	AppPackage    = model.AppPackage
	Theme         = model.Theme
	SystemConfig  = model.SystemConfig
	WindowID      = model.WindowID
	Rect          = geom.Rect
	Buffer        = model.Buffer
	InputEvent    = model.InputEvent
	PointerAction = model.PointerAction
	PointerSource = model.PointerSource
	PointerSample = model.PointerSample
	PointerEvent  = model.PointerEvent
	KeyAction     = model.KeyAction
	KeyEvent      = model.KeyEvent
	TextEvent     = model.TextEvent
	ScrollEvent   = model.ScrollEvent
	SystemAction  = model.SystemAction
	SystemEvent   = model.SystemEvent
	RotateEvent   = model.RotateEvent
)

const (
	LightTheme      = model.LightTheme
	DarkTheme       = model.DarkTheme
	PointerDown     = model.PointerDown
	PointerMove     = model.PointerMove
	PointerUp       = model.PointerUp
	PointerCancel   = model.PointerCancel
	TouchSource     = model.TouchSource
	StylusSource    = model.StylusSource
	MouseSource     = model.MouseSource
	TrackpadSource  = model.TrackpadSource
	KeyDown         = model.KeyDown
	KeyUp           = model.KeyUp
	SystemBack      = model.SystemBack
	SystemHome      = model.SystemHome
	SystemUnlock    = model.SystemUnlock
	SystemSettings  = model.SystemSettings
	SystemScreenOff = model.SystemScreenOff
	SystemScreenOn  = model.SystemScreenOn
	SystemNotices   = model.SystemNotices
	SystemRecents   = model.SystemRecents
)

func DefaultSystemConfig() SystemConfig { return model.DefaultSystemConfig() }
