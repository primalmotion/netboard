//go:build !cgo

package cboard

import "fmt"

// / NewLibClipboardManager returns an error.
func NewLibClipboardManager() (ClipboardManager, error) {
	return nil, fmt.Errorf("lib mode is not supported on this platform. try another mode")
}
