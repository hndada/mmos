package model

import "mmos/internal/common/geom"

type WindowID string

type Buffer struct {
	Revision int
	Content  string
	Bounds   geom.Rect
}
