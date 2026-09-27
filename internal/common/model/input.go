package model

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

type SystemAction uint8

const (
	SystemBack SystemAction = iota
	SystemHome
	SystemUnlock
	SystemSettings
	SystemScreenOff
	SystemScreenOn
	SystemNotices
	SystemRecents
)

type SystemEvent struct{ Action SystemAction }

func (SystemEvent) isInputEvent() {}

// RotateEvent requests a server-owned display orientation change.
type RotateEvent struct{}

func (RotateEvent) isInputEvent() {}
