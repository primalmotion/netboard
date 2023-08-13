package cboard

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

const ClipboardEmptyErrorString = "Nothing is copied\n"

type toolsClipboardManager struct {
}

// NewToolsClipboardManager returns a new ClipboardManager
// using wl-clipboard underneath.
func NewToolsClipboardManager() (ClipboardManager, error) {

	if _, err := exec.LookPath("wl-copy"); err != nil {
		return nil, fmt.Errorf("unable to find wl-copy binary: either install wl-clipboard or try another mode")
	}

	if _, err := exec.LookPath("wl-paste"); err != nil {
		return nil, fmt.Errorf("unable to find wl-paste binary: either install wl-clipboard or try another mode")
	}

	return &toolsClipboardManager{}, nil
}

func (c *toolsClipboardManager) Read() ([]byte, error) {

	cmd := exec.Command("wl-paste", "--no-newline")

	stdout := bytes.NewBuffer(nil)
	cmd.Stdout = stdout
	stderr := bytes.NewBuffer(nil)
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		if stderr.String() == ClipboardEmptyErrorString {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to run read command: %w", err)
	}

	return stdout.Bytes(), nil
}

func (c *toolsClipboardManager) Write(data []byte) error {

	cmd := exec.Command("wl-copy", "--trim-newline")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("unable to acquire stdin pipe: %w", err)
	}

	if _, err := io.Copy(stdin, bytes.NewBuffer(data)); err != nil {
		return fmt.Errorf("unable to retrieve data from stdin pipe: %w", err)
	}
	stdin.Close() //nolint

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to wait command: %w", err)
	}

	return nil
}

func (c *toolsClipboardManager) Watch(ctx context.Context) (<-chan []byte, <-chan error) {

	chout := make(chan []byte)
	cherr := make(chan error)

	go func() {

		cmd := exec.Command("wl-paste", "--no-newline", "-w", "echo")

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			cherr <- fmt.Errorf("unable to bind stdout: %w", err)
			return
		}
		defer stdout.Close() //nolint

		stderr, err := cmd.StderrPipe()
		if err != nil {
			cherr <- fmt.Errorf("unable to bind stderr: %w", err)
			return
		}
		defer stderr.Close() //nolint

		buf := bytes.NewBuffer(nil)
		go func() {
			io.Copy(buf, stderr)
		}()

		if err := cmd.Start(); err != nil {
			cherr <- fmt.Errorf("unable to start command: %w", err)
			return
		}

		scan := bufio.NewScanner(stdout)

		go func() {
			for scan.Scan() {
				data, err := c.Read()
				if err != nil {
					cherr <- fmt.Errorf("unable to scan stdout: %w", err)
					return
				}

				if len(data) <= 0 {
					continue
				}

				select {
				case chout <- data:
				case <-ctx.Done():
				default:
				}
			}
		}()
		if err := cmd.Wait(); err != nil {
			cherr <- fmt.Errorf("error while listening wl-paste: %w", err)
		}
	}()

	return chout, cherr
}
