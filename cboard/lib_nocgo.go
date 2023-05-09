//go:build !cgo

package cboard

import "log"

// / NewLibClipboardManager returns an error.
func NewLibClipboardManager() ClipboardManager {
	log.Fatal("lib mode is not supported on this platform. Try another mode")
	return nil
}
