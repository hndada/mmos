package settings

import (
	"testing"

	"mmos/internal/common/model"
)

func TestHandleInputTogglesThemeThroughApply(t *testing.T) {
	var applied model.SystemConfig
	var app *App
	app = New(&model.AppPackage{ID: "settings"}, func(config model.SystemConfig) bool {
		applied = config
		app.Configure(config)
		return true
	})
	app.Configure(model.DefaultSystemConfig())

	if !app.HandleInput("settings", tap(160, 120)) {
		t.Fatal("toggle theme")
	}
	if applied.Theme != model.DarkTheme || app.Process.Windows["settings"].Node("theme").Text != "Theme: Dark" {
		t.Fatalf("config=%#v text=%q", applied, app.Process.Windows["settings"].Node("theme").Text)
	}
}

func TestHandleInputTogglesLockOnScreenOff(t *testing.T) {
	var applied model.SystemConfig
	var app *App
	app = New(&model.AppPackage{ID: "settings"}, func(config model.SystemConfig) bool {
		applied = config
		app.Configure(config)
		return true
	})
	app.Configure(model.DefaultSystemConfig())

	if !app.HandleInput("settings", tap(160, 200)) {
		t.Fatal("toggle lock")
	}
	if !applied.LockOnScreenOff || app.Process.Windows["settings"].Node("lock").Text != "Lock after screen off: On" {
		t.Fatalf("config=%#v text=%q", applied, app.Process.Windows["settings"].Node("lock").Text)
	}
}

func tap(x, y int) model.PointerEvent {
	return model.PointerEvent{
		Source: model.TouchSource, Action: model.PointerUp, ChangedPointerID: 0,
		Pointers: []model.PointerSample{{ID: 0, X: x, Y: y}},
	}
}
