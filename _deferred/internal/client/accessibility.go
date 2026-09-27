package client

// Accessibility provides the runtime traversal used by a screen reader or a
// single-switch controller. It works only from app-provided semantic nodes;
// the window server never inspects an application's UI tree.
type Accessibility struct{ index int }

// Next moves to the next focusable node and returns its accessible label.
func (a *Accessibility) Next(window *Window) (string, bool) {
	if window == nil {
		return "", false
	}
	nodes := window.FocusOrder()
	if len(nodes) == 0 {
		return "", false
	}
	if a.index >= len(nodes) {
		a.index = 0
	}
	node := nodes[a.index]
	a.index = (a.index + 1) % len(nodes)
	return node.Label, true
}

// Activate performs the current switch action through the normal command
// path. Callers still decide which app command is permitted.
func (a *Accessibility) Activate(window *Window) (Command, bool) {
	if window == nil {
		return "", false
	}
	nodes := window.FocusOrder()
	if len(nodes) == 0 {
		return "", false
	}
	index := a.index - 1
	if index < 0 {
		index = len(nodes) - 1
	}
	return nodes[index].Command, true
}
