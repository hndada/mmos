package home

import (
	"testing"

	"mmos/internal/common/model"
	"mmos/internal/protocol"
)

type launcherFunc func(protocol.LaunchRequest) protocol.LaunchReply

func (f launcherFunc) Launch(request protocol.LaunchRequest) protocol.LaunchReply {
	return f(request)
}

func TestHandleInputLaunchesSelectedApp(t *testing.T) {
	launched := false
	app := New(&model.AppPackage{ID: "home"}, launcherFunc(func(request protocol.LaunchRequest) protocol.LaunchReply {
		if request.PackageID != "chat-app" {
			t.Fatalf("package = %q", request.PackageID)
		}
		launched = true
		return protocol.LaunchReply{Started: true}
	}))

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
