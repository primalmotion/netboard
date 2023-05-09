package cboard

import (
	"bufio"
	"context"
	"io"
	"os/exec"
)

type toolsClipboardManager struct {
}

func NewToolsClipboardManager() ClipboardManager {
	return &toolsClipboardManager{}
}

func (c *toolsClipboardManager) Read() ([]byte, error) {

	cmd := exec.Command("wl-paste", "--no-newline")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	return data, cmd.Wait()
}

func (c *toolsClipboardManager) Write(data []byte) error {

	cmd := exec.Command("wl-copy", "--foreground", "--trim-newline")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := stdin.Write(data); err != nil {
		return err
	}
	stdin.Close()

	return cmd.Wait()
}

func (c *toolsClipboardManager) Watch(ctx context.Context) <-chan []byte {

	ch := make(chan []byte)

	go func() {

		cmd := exec.Command("wl-paste", "-w", "echo")

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}

		if err := cmd.Start(); err != nil {
			panic(err)
		}

		scan := bufio.NewScanner(stdout)

		for scan.Scan() {
			data, err := c.Read()
			if err != nil {
				panic(err)
			}

			select {
			case ch <- data:
			case <-ctx.Done():
			default:
			}
		}
	}()

	return ch
}
