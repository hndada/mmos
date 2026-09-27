package server

import "errors"

type Permission uint8

const (
	Camera Permission = iota
	Microphone
	Location
	Files
	Clipboard
	ScreenCapture
)

type PermissionState uint8

const (
	PermissionAsk PermissionState = iota
	PermissionGranted
	PermissionDenied
)

type PermissionRequest struct {
	ID         int
	PackageID  string
	Permission Permission
}

var ErrPermissionRequestMissing = errors.New("permission request not found")

type permissions struct {
	nextID   int
	pending  map[int]PermissionRequest
	decision map[string]map[Permission]PermissionState
}

func newPermissions() permissions {
	return permissions{pending: map[int]PermissionRequest{}, decision: map[string]map[Permission]PermissionState{}}
}

func (r *Runtime) RequestPermission(s Session, permission Permission) (PermissionRequest, PermissionState, bool) {
	if !r.authorized(s) {
		return PermissionRequest{}, PermissionAsk, false
	}
	process, ok := r.processes.processes[s.pid]
	if !ok {
		return PermissionRequest{}, PermissionAsk, false
	}
	if state := r.permissions.state(process.PackageID, permission); state != PermissionAsk {
		return PermissionRequest{}, state, true
	}
	r.permissions.nextID++
	request := PermissionRequest{ID: r.permissions.nextID, PackageID: process.PackageID, Permission: permission}
	r.permissions.pending[request.ID] = request
	return request, PermissionAsk, true
}

func (r *Runtime) ResolvePermission(id int, granted bool) (PermissionRequest, error) {
	request, ok := r.permissions.pending[id]
	if !ok {
		return PermissionRequest{}, ErrPermissionRequestMissing
	}
	state := PermissionDenied
	if granted {
		state = PermissionGranted
	}
	r.permissions.set(request.PackageID, request.Permission, state)
	delete(r.permissions.pending, id)
	return request, nil
}

func (r *Runtime) PermissionState(s Session, permission Permission) (PermissionState, bool) {
	if !r.authorized(s) {
		return PermissionAsk, false
	}
	process, ok := r.processes.processes[s.pid]
	if !ok {
		return PermissionAsk, false
	}
	return r.permissions.state(process.PackageID, permission), true
}

// RevokePermission is trusted system policy. Apps can request a capability but
// cannot grant or revoke decisions for any package.
func (r *Runtime) RevokePermission(packageID string, permission Permission) bool {
	if _, ok := r.packages[packageID]; !ok {
		return false
	}
	r.permissions.set(packageID, permission, PermissionAsk)
	return true
}

func (p *permissions) state(packageID string, permission Permission) PermissionState {
	return p.decision[packageID][permission]
}
func (p *permissions) set(packageID string, permission Permission, state PermissionState) {
	if p.decision[packageID] == nil {
		p.decision[packageID] = map[Permission]PermissionState{}
	}
	p.decision[packageID][permission] = state
}
