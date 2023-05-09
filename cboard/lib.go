package cboard

import (
	"context"

	"golang.design/x/clipboard"
)

type libClipboardManager struct {
}

// NewLibClipboardManager returns a new ClipboardManager
// base on golang.design/x/clipboard
func NewLibClipboardManager() ClipboardManager {
	return &libClipboardManager{}
}

func (c *libClipboardManager) Read() ([]byte, error) {
	return clipboard.Read(clipboard.FmtText), nil
}

func (c *libClipboardManager) Write(data []byte) error {
	clipboard.Write(clipboard.FmtText, data)
	return nil
}

func (c *libClipboardManager) Watch(ctx context.Context) <-chan []byte {
	return clipboard.Watch(ctx, clipboard.FmtText)
}
