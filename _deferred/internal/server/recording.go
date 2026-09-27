package server

import "errors"

var ErrRecordingActive = errors.New("recording already active")

// Recording is a trusted system-owned capture session. Its frames use the
// same protected-window filtering as a screenshot.
type Recording struct {
	DisplayID DisplayID
	Frames    []Frame
	Active    bool
}

// StartRecording begins a system capture. Applications never receive a
// recording handle and therefore cannot capture another app's display.
func (r *Runtime) StartRecording(displayID DisplayID) error {
	if _, ok := r.displays[displayID]; !ok {
		return ErrDisplayNotFound
	}
	if recording := r.recordings[displayID]; recording.Active {
		return ErrRecordingActive
	}
	r.recordings[displayID] = Recording{DisplayID: displayID, Active: true}
	return nil
}

// StopRecording returns a copy of the trusted recording and stops capture.
func (r *Runtime) StopRecording(displayID DisplayID) (Recording, bool) {
	recording, ok := r.recordings[displayID]
	if !ok || !recording.Active {
		return Recording{}, false
	}
	recording.Active = false
	recording.Frames = append([]Frame(nil), recording.Frames...)
	r.recordings[displayID] = recording
	return recording, true
}

func (r *Runtime) record(displayID DisplayID) {
	recording := r.recordings[displayID]
	if !recording.Active {
		return
	}
	frame, ok := r.Capture(displayID)
	if !ok {
		return
	}
	recording.Frames = append(recording.Frames, frame)
	r.recordings[displayID] = recording
}
