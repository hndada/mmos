package client

import "testing"

func TestAccessibilityTraversesSemanticFocusOrder(t *testing.T) {
	window := NewWindow("test", "root")
	window.Root.Children = []*UINode{
		{ID: "disabled", Visible: true, Enabled: false, Command: "skip", Label: "Skip"},
		{ID: "send", Visible: true, Enabled: true, Command: "send", Label: "Send message"},
	}
	var accessibility Accessibility
	if label, ok := accessibility.Next(window); !ok || label != "Send message" {
		t.Fatalf("next = %q found=%t", label, ok)
	}
	if command, ok := accessibility.Activate(window); !ok || command != "send" {
		t.Fatalf("activate = %q found=%t", command, ok)
	}
}
