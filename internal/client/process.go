package client

import "mmos/internal/common/model"

// AppProcess is the client-side execution container for one app. It owns
// application state, the UINode hierarchy, and the buffers it draws.
type AppProcess struct {
	Package *model.AppPackage
	Windows map[model.WindowID]*Window
	Config  model.SystemConfig
}
