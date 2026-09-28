package server

import "time"

// motion is a display-local slide of one trusted system surface.
type motion struct {
	windowID WindowID
	offsetY  int
	started  time.Time
	duration time.Duration
}

func (m *motion) start(id WindowID, offsetY int, duration time.Duration) {
	m.windowID = id
	m.offsetY = offsetY
	m.started = time.Now()
	m.duration = duration
}

func (m *motion) apply(layers []Layer, now time.Time) []Layer {
	if m.started.IsZero() {
		return layers
	}
	progress := ease(now.Sub(m.started), m.duration)
	if progress >= 1 {
		*m = motion{}
		return layers
	}
	for i := range layers {
		if layers[i].WindowID == m.windowID {
			layers[i].OffsetY = int(float64(m.offsetY) * (1 - progress))
			return layers
		}
	}
	*m = motion{}
	return layers
}
