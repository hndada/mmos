package model

type Theme uint8

const (
	LightTheme Theme = iota
	DarkTheme
)

type SystemConfig struct {
	Revision          uint64
	Theme             Theme
	FontScale         float32
	ReducedMotion     bool
	Locale            string
	OrientationLocked bool
	// LockOnScreenOff requires unlock after the display is turned off.
	LockOnScreenOff bool
}

func DefaultSystemConfig() SystemConfig {
	return SystemConfig{Revision: 1, Theme: LightTheme, FontScale: 1, Locale: "en-US"}
}
