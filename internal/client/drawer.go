package client

import "fmt"

// Drawer is run by the client process and draws client UI into an app-owned
// buffer. Only the resulting Buffer crosses to the model.
type Drawer struct{}

func (Drawer) Draw(window *Window) {
	window.Buffer.Revision++
	window.Buffer.Bounds = window.Bounds()
	window.Buffer.Content = fmt.Sprintf(
		"%s(root=%s, %dx%d)",
		window.ID,
		window.Root.ID,
		window.Buffer.Bounds.Width,
		window.Buffer.Bounds.Height,
	)
}
