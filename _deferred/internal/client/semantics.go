package client

// Role and the following fields were the 3권 semantic-node extension.
type Role uint8

const (
	Group Role = iota
	StaticText
	ButtonRole
)

// SemanticNode shows the fields formerly added to UINode.
type SemanticNode struct {
	Role  Role
	Label string
}

func (n *UINode) Focusable() bool {
	return n != nil && n.Visible && n.Enabled && n.Command != ""
}

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
