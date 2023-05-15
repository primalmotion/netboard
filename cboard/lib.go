//go:build cgo

package cboard

import (
	"context"
	"fmt"

	"golang.design/x/clipboard"
)

type libClipboardManager struct {
}

// NewLibClipboardManager returns a new ClipboardManager
// base on golang.design/x/clipboard
func NewLibClipboardManager() (ClipboardManager, error) {

	if err := clipboard.Init(); err != nil {
		return nil, fmt.Errorf("unable to initialize clipboard: %w", err)
	}

	return &libClipboardManager{}, nil
}

func (c *libClipboardManager) Read() ([]byte, error) {
	return clipboard.Read(clipboard.FmtText), nil
}

func (c *libClipboardManager) Write(data []byte) error {
	clipboard.Write(clipboard.FmtText, data)
	return nil
}

func (c *libClipboardManager) Watch(ctx context.Context) (<-chan []byte, <-chan error) {
	chout := clipboard.Watch(ctx, clipboard.FmtText)
	return chout, make(chan error)
}
