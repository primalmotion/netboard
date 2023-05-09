package cboard

import "context"

// A ClipboardManager is the interface describing an
// object that can manipulates the OS clibboard.
type ClipboardManager interface {
	Read() ([]byte, error)
	Write([]byte) error
	Watch(context.Context) <-chan []byte
}
