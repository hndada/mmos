package server

import (
	"testing"

	"mmos/internal/common/model"
)

func TestRuntimeOwnsNoticeHistory(t *testing.T) {
	runtime := NewRuntime(&model.AppPackage{ID: "chat", EntryPoint: "Chat.Main"})
	launch, err := runtime.Launch("chat")
	if err != nil {
		t.Fatal(err)
	}
	notice, ok := runtime.PublishNotice(launch.Session, "Chat", "Message sent")
	if !ok || notice.PackageID != "chat" || len(runtime.Notices()) != 1 {
		t.Fatalf("notice = %#v found=%t", notice, ok)
	}
	if !runtime.DismissNotice(notice.ID) || len(runtime.Notices()) != 0 {
		t.Fatal("dismiss notice")
	}
}
