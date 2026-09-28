// Package ime implements the trusted on-screen input method surface.
package ime

import (
	"errors"
	"fmt"
	"time"

	"mmos/internal/common/geom"
	"mmos/internal/server"
)

var ErrInstall = errors.New("install ime")

const window server.WindowID = "ime"

const animationDuration = 180 * time.Millisecond

// App owns the system keyboard and turns its touches into text input. Client
// apps only receive TextEvent; they never receive touches on this overlay.
type App struct {
	runtime *server.Runtime
	display server.DisplayID
	visible bool
	target  server.WindowID
	raw     []rune
}

func New(runtime *server.Runtime, display server.DisplayID) (*App, error) {
	if runtime == nil {
		return nil, ErrInstall
	}
	config, ok := runtime.Display(display)
	if !ok || !runtime.RegisterSystemWindow(window, display, keyboardBounds(config.Bounds())) ||
		!runtime.HideSystemWindow(window) {
		return nil, ErrInstall
	}
	return &App{runtime: runtime, display: display}, nil
}

func (a *App) Visible() bool { return a.visible }

func (a *App) Show(target server.WindowID) (server.Frame, bool) {
	if target == "" {
		return a.runtime.LastFrame(a.display), false
	}
	config, ok := a.runtime.Display(a.display)
	if !ok || !a.runtime.SubmitSystem(window, server.Buffer{
		Revision: config.Revision,
		Bounds:   keyboardBounds(config.Bounds()),
		Content:  a.content(),
	}) || !a.runtime.ShowSystemWindow(window) ||
		!a.runtime.SlideSystemWindow(window, keyboardBounds(config.Bounds()).Height, animationDuration) {
		return a.runtime.LastFrame(a.display), false
	}
	a.visible = true
	a.target = target
	return a.runtime.VSync(a.display), true
}

func (a *App) Hide() (server.Frame, bool) {
	if !a.visible || !a.runtime.HideSystemWindow(window) {
		return a.runtime.LastFrame(a.display), false
	}
	a.visible = false
	a.target = ""
	a.raw = nil
	return a.runtime.VSync(a.display), true
}

func (a *App) Target() server.WindowID { return a.target }

func (a *App) Inset() int {
	config, ok := a.runtime.Display(a.display)
	if !ok {
		return 0
	}
	return keyboardBounds(config.Bounds()).Height
}

// HandleInput consumes only touches inside the visible keyboard. The returned
// event is routed by composition to the window selected when it was shown.
func (a *App) HandleInput(event server.InputEvent) (server.TextEvent, bool) {
	if !a.visible {
		return server.TextEvent{}, false
	}
	p, ok := event.(server.PointerEvent)
	if !ok || p.Action != server.PointerUp || p.ChangedPointerID != 0 {
		return server.TextEvent{}, false
	}
	sample, ok := p.PointerByID(0)
	if !ok {
		return server.TextEvent{}, false
	}
	config, ok := a.runtime.Display(a.display)
	if !ok || !keyboardBounds(config.Bounds()).Contains(sample.X, sample.Y) {
		return server.TextEvent{}, false
	}

	switch keyAt(sample.X, sample.Y, config.Bounds()) {
	case "backspace":
		if len(a.raw) > 0 {
			a.raw = a.raw[:len(a.raw)-1]
		}
	case "commit":
		text := compose(a.raw)
		a.raw = nil
		a.refresh()
		return server.TextEvent{Text: text}, true
	default:
		a.raw = append(a.raw, []rune(keyAt(sample.X, sample.Y, config.Bounds()))...)
	}
	a.refresh()
	return server.TextEvent{Text: compose(a.raw), Composing: true}, true
}

func (a *App) Resize() bool {
	config, ok := a.runtime.Display(a.display)
	if !ok || !a.runtime.ResizeSystemWindow(window, keyboardBounds(config.Bounds())) {
		return false
	}
	return !a.visible || a.refresh()
}

func (a *App) refresh() bool {
	if !a.visible {
		return true
	}
	config, ok := a.runtime.Display(a.display)
	return ok && a.runtime.SubmitSystem(window, server.Buffer{
		Revision: config.Revision,
		Bounds:   keyboardBounds(config.Bounds()),
		Content:  a.content(),
	})
}

func (a *App) content() string {
	return fmt.Sprintf("ime(preedit=%q, rows=[ㅂㅈㄷㄱㅅㅛㅕㅑㅐㅔ ㅁㄴㅇㄹㅎㅗㅓㅏㅣ], backspace, commit)", compose(a.raw))
}

func keyboardBounds(display geom.Rect) geom.Rect {
	const height = 144
	return geom.Rect{Y: display.Height - height, Width: display.Width, Height: height}
}

func keyAt(x, y int, display geom.Rect) string {
	bounds := keyboardBounds(display)
	row := (y - bounds.Y) * 3 / bounds.Height
	column := (x - bounds.X) * 10 / bounds.Width
	switch row {
	case 0:
		return []string{"ㅂ", "ㅈ", "ㄷ", "ㄱ", "ㅅ", "ㅛ", "ㅕ", "ㅑ", "ㅐ", "ㅔ"}[column]
	case 1:
		if column < 9 {
			return []string{"ㅁ", "ㄴ", "ㅇ", "ㄹ", "ㅎ", "ㅗ", "ㅓ", "ㅏ", "ㅣ"}[column]
		}
		return "backspace"
	default:
		if column < 5 {
			return "backspace"
		}
		return "commit"
	}
}

func compose(raw []rune) string {
	var out []rune
	for i := 0; i < len(raw); {
		if i+1 < len(raw) {
			final := 0
			used := 2
			if i+2 < len(raw) && (i+3 >= len(raw) || !isVowel(raw[i+3])) {
				if index, ok := finalIndex(raw[i+2]); ok {
					final, used = index, 3
				}
			}
			if syllable, ok := hangul(raw[i], raw[i+1], final); ok {
				out = append(out, syllable)
				i += used
				continue
			}
		}
		out = append(out, raw[i])
		i++
	}
	return string(out)
}

func hangul(initial, vowel rune, final int) (rune, bool) {
	initials := []rune("ㄱㄲㄴㄷㄸㄹㅁㅂㅃㅅㅆㅇㅈㅉㅊㅋㅌㅍㅎ")
	vowels := []rune("ㅏㅐㅑㅒㅓㅔㅕㅖㅗㅘㅙㅚㅛㅜㅝㅞㅟㅠㅡㅢㅣ")
	initialIndex, vowelIndex := -1, -1
	for i, r := range initials {
		if r == initial {
			initialIndex = i
		}
	}
	for i, r := range vowels {
		if r == vowel {
			vowelIndex = i
		}
	}
	if initialIndex < 0 || vowelIndex < 0 {
		return 0, false
	}
	return rune(0xAC00 + (initialIndex*21+vowelIndex)*28 + final), true
}

func isVowel(r rune) bool {
	for _, vowel := range []rune("ㅏㅐㅑㅒㅓㅔㅕㅖㅗㅘㅙㅚㅛㅜㅝㅞㅟㅠㅡㅢㅣ") {
		if r == vowel {
			return true
		}
	}
	return false
}

func finalIndex(r rune) (int, bool) {
	for index, consonant := range []rune("ㄱㄲㄳㄴㄵㄶㄷㄹㄺㄻㄼㄽㄾㄿㅀㅁㅂㅄㅅㅆㅇㅈㅊㅋㅌㅍㅎ") {
		if r == consonant {
			return index + 1, true
		}
	}
	return 0, false
}
