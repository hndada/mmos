package client

import "mmos/internal/common/geom"

type NodeKind uint8

const (
	Container NodeKind = iota
	Text
	Button
)

// Role describes how assistive technology should announce a node. It stays
// separate from Kind because a client may draw the same kind of node with a
// different semantic purpose.
type Role uint8

const (
	Group Role = iota
	StaticText
	ButtonRole
)

type Command string

// UINode is client-owned application UI state. The server never changes it.
type UINode struct {
	ID       string
	Kind     NodeKind
	Bounds   geom.Rect
	Text     string
	Role     Role
	Label    string
	Visible  bool
	Enabled  bool // Only read for nodes that expose a Command.
	Command  Command
	Children []*UINode
}

func (n *UINode) Focusable() bool {
	return n != nil && n.Visible && n.Enabled && n.Command != ""
}

// FocusOrder appends this visible subtree's actionable nodes in document
// order. The client owns this traversal; the server never examines UI nodes.
func (n *UINode) FocusOrder(nodes []*UINode) []*UINode {
	if n == nil || !n.Visible {
		return nodes
	}
	if n.Focusable() {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children {
		nodes = child.FocusOrder(nodes)
	}
	return nodes
}

func (n *UINode) HitTest(x, y int) *UINode {
	if n == nil || !n.Visible || !n.Bounds.Contains(x, y) {
		return nil
	}
	for index := len(n.Children) - 1; index >= 0; index-- {
		if hit := n.Children[index].HitTest(x, y); hit != nil {
			return hit
		}
	}
	return n
}

func (n *UINode) Find(id string) *UINode {
	if n == nil {
		return nil
	}
	if n.ID == id {
		return n
	}
	for _, child := range n.Children {
		if found := child.Find(id); found != nil {
			return found
		}
	}
	return nil
}
