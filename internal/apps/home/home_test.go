package home

import (
	"testing"

	"mmos/internal/common/model"
)

func TestHandleInputLaunchesSelectedApp(t *testing.T) {
	launched := false
	app := New(&model.AppPackage{ID: "home"}, func() bool {
		launched = true
		return true
	})

	changed := app.HandleInput("home", model.PointerEvent{
		Source:           model.TouchSource,
		Action:           model.PointerUp,
		ChangedPointerID: 0,
		Pointers:         []model.PointerSample{{ID: 0, X: 160, Y: 210}},
	})
	if !changed || !launched {
		t.Fatalf("launch = changed:%t launched:%t", changed, launched)
	}
}
