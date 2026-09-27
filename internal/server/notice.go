package server

// Notice is a system-owned presentation of an app event. It intentionally
// carries plain content only; the app cannot control the trusted overlay.
type Notice struct {
	ID        int
	PackageID string
	Title     string
	Body      string
}

func (r *Runtime) PublishNotice(s Session, title, body string) (Notice, bool) {
	if !r.authorized(s) {
		return Notice{}, false
	}
	process, ok := r.processes.processes[s.pid]
	if !ok {
		return Notice{}, false
	}
	r.nextNotice++
	notice := Notice{ID: r.nextNotice, PackageID: process.PackageID, Title: title, Body: body}
	r.notices = append(r.notices, notice)
	return notice, true
}

// Notices returns a snapshot for trusted system UI.
func (r *Runtime) Notices() []Notice { return append([]Notice(nil), r.notices...) }

func (r *Runtime) DismissNotice(id int) bool {
	for i, notice := range r.notices {
		if notice.ID == id {
			r.notices = append(r.notices[:i], r.notices[i+1:]...)
			return true
		}
	}
	return false
}
