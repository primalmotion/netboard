package cboard

import (
	"bufio"
	"context"
	"fmt"
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
		return nil, fmt.Errorf("unable to init command: %w", err)
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("unable to start command: %w", err)
	}

	data, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("unable to read stdin: %w", err)
	}

	cmd.Wait()

	return data, nil
}

func (c *toolsClipboardManager) Write(data []byte) error {

	cmd := exec.Command("wl-copy", "--foreground", "--trim-newline")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("unable to init command: %w", err)
	}
	defer stdin.Close()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start command: %w", err)
	}

	if _, err := stdin.Write(data); err != nil {
		return fmt.Errorf("unable to write stdin: %w", err)
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
