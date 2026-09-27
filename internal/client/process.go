package client

import "mmos/internal/common/model"

// AppProcess is the client-side execution container for one app. It owns
// application state, the UINode hierarchy, and the buffers it draws.
type AppProcess struct {
	Package *model.AppPackage
	Windows map[model.WindowID]*Window
	Config  model.SystemConfig
}

// Configure records the latest system configuration before the application
// lays out its UI and submits another buffer.
func (p *AppProcess) Configure(config model.SystemConfig) {
	p.Config = config
}
