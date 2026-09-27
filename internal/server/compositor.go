package server

type Layer struct {
	WindowID WindowID
	Buffer   Buffer
	Opacity  float64
	ZIndex   int
}

type Frame struct {
	Number    int
	DisplayID DisplayID
	Display   DisplayConfig
	Layers    []Layer
}

// Presents reports whether the compositor adopted this buffer revision into
// the frame. Drawing or submitting alone does not make a buffer visible.
func (f Frame) Presents(id WindowID, revision int) bool {
	for _, layer := range f.Layers {
		if layer.WindowID == id && layer.Buffer.Revision == revision {
			return true
		}
	}
	return false
}

// Compositor combines submitted client buffers in server z-order.
type Compositor struct{ LastFrame Frame }

func (c *Compositor) Compose(id DisplayID, server WindowServer, display DisplayConfig) Frame {
	layers := make([]Layer, 0, len(server.ZOrder))
	for _, id := range server.ZOrder {
		window, ok := server.Windows[id]
		if ok && window.Visible {
			layers = append(layers, Layer{
				WindowID: id,
				Buffer:   window.Buffer,
				Opacity:  1,
				ZIndex:   len(layers),
			})
		}
	}
	c.LastFrame = Frame{
		Number:    c.LastFrame.Number + 1,
		DisplayID: id,
		Display:   display,
		Layers:    layers,
	}
	return c.LastFrame
}

// Capture returns the most recently composed frame without excluded windows.
// The caller owns the policy that decides which windows to exclude.
func (c Compositor) Capture(excluded map[WindowID]bool) Frame {
	capture := c.LastFrame
	capture.Layers = make([]Layer, 0, len(c.LastFrame.Layers))
	for _, layer := range c.LastFrame.Layers {
		if excluded[layer.WindowID] {
			continue
		}
		layer.ZIndex = len(capture.Layers)
		capture.Layers = append(capture.Layers, layer)
	}
	return capture
}
