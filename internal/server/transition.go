package server

import "time"

const transitionDuration = 180 * time.Millisecond

// transition retains the previous buffer while the newly focused window fades
// in. It is scoped to one display and never changes window ownership.
type transition struct {
	from    Layer
	to      WindowID
	started time.Time
}

func (t *transition) start(from *Window, to WindowID) {
	if from == nil || from.ID == to {
		t.clear()
		return
	}
	t.from = Layer{WindowID: from.ID, Buffer: from.Buffer, Opacity: 1}
	t.to = to
	t.started = time.Now()
}

func (t *transition) clear()       { *t = transition{} }
func (t *transition) active() bool { return !t.started.IsZero() }

func (t *transition) apply(layers []Layer, now time.Time) []Layer {
	if !t.active() {
		return layers
	}
	progress := ease(now.Sub(t.started), transitionDuration)
	if progress >= 1 {
		t.clear()
		return layers
	}
	for i := range layers {
		if layers[i].WindowID != t.to {
			continue
		}
		layers[i].Opacity = progress
		from := t.from
		from.Opacity = 1 - progress
		for j := range layers {
			if layers[j].WindowID == from.WindowID {
				layers[j].Opacity = from.Opacity
				return layers
			}
		}
		layers = append(layers, Layer{})
		copy(layers[i+1:], layers[i:])
		layers[i] = from
		for j := range layers {
			layers[j].ZIndex = j
		}
		return layers
	}
	t.clear()
	return layers
}

func ease(elapsed, duration time.Duration) float64 {
	if elapsed <= 0 {
		return 0
	}
	if elapsed >= duration {
		return 1
	}
	p := float64(elapsed) / float64(duration)
	return p * p * (3 - 2*p)
}
