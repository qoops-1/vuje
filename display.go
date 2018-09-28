package main

import (
	"github.com/bitly/go-simplejson"
)

// Display represents a currently visible document with all representational parameters
type Display struct {
	Doc              *simplejson.Json
	DocHeight        int
	DocOffsetY       int
	ActiveCompletion int
	OnlyKeys         bool
}

// MoveWindow moves DocOffsetY of the display by up to "y" lines, so that
// the window remains within the boundaries of the document (+1 trailing line)
func (dspl *Display) MoveWindow(y int, windowSize int) bool {
	contentsHeight := windowSize - (completionY + 1)
	moveUpWhenTop := y < 0 && dspl.DocOffsetY == 0
	moveDownWhenBottom := y > 0 && (dspl.DocOffsetY+contentsHeight == dspl.DocHeight+1)
	if moveUpWhenTop || moveDownWhenBottom || y == 0 {
		return false
	}
	newOffset := dspl.DocOffsetY + y
	maxOffset := dspl.DocHeight - contentsHeight + 1
	if newOffset <= 0 {
		dspl.DocOffsetY = 0
	} else if newOffset >= maxOffset {
		dspl.DocOffsetY = maxOffset
	} else {
		dspl.DocOffsetY = newOffset
	}
	return true
}
