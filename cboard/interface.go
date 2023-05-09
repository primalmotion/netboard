package cboard

import "context"

type ClipboardManager interface {
	Read() ([]byte, error)
	Write([]byte) error
	Watch(context.Context) <-chan []byte
}
