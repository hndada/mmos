package server

import "mmos/internal/common/model"

// ResourceLimit is a trusted per-process budget. Zero means unlimited for
// that resource, which keeps the base POC lightweight until policy enables a
// limit.
type ResourceLimit struct {
	MemoryBytes int
	CPUUnits    int
	GPUUnits    int
}

type resourceUse struct {
	CPUUnits int
	GPUUnits int
}

func (r *Runtime) SetResourceLimit(packageID string, limit ResourceLimit) bool {
	if _, ok := r.packages[packageID]; !ok || limit.MemoryBytes < 0 || limit.CPUUnits < 0 || limit.GPUUnits < 0 {
		return false
	}
	r.limits[packageID] = limit
	return true
}

// Charge records CPU/GPU work for the current accounting interval. It rejects
// excess work instead of silently allowing an app to exceed its quota.
func (r *Runtime) Charge(s Session, cpu, gpu int) bool {
	if !r.authorized(s) || cpu < 0 || gpu < 0 {
		return false
	}
	process := r.processes.processes[s.pid]
	limit := r.limits[process.PackageID]
	use := r.usage[s.pid]
	if limit.CPUUnits > 0 && use.CPUUnits+cpu > limit.CPUUnits ||
		limit.GPUUnits > 0 && use.GPUUnits+gpu > limit.GPUUnits {
		return false
	}
	use.CPUUnits += cpu
	use.GPUUnits += gpu
	r.usage[s.pid] = use
	return true
}

// ResetUsage begins a new trusted accounting interval.
func (r *Runtime) ResetUsage() { r.usage = map[int]resourceUse{} }

func (r *Runtime) acceptsBuffer(s Session, buffer model.Buffer) bool {
	process := r.processes.processes[s.pid]
	limit := r.limits[process.PackageID]
	if limit.MemoryBytes == 0 {
		return true
	}
	if buffer.Bounds.Width < 0 || buffer.Bounds.Height < 0 {
		return false
	}
	return int64(buffer.Bounds.Width)*int64(buffer.Bounds.Height)*4 <= int64(limit.MemoryBytes)
}
