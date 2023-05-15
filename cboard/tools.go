package cboard

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

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
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("unable to bind stdout: %w", err)
	}
	defer stdout.Close() // nolint

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("unable to bind stderr: %w", err)
	}
	defer stderr.Close() //nolint

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("unable to start command: %w", err)
	}

	data, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("unable to read stdout: %w", err)
	}

	stderrData, err := io.ReadAll(stderr)
	if err != nil {
		return nil, fmt.Errorf("unable to read stderr: %w", err)
	}
	stderr.Close() // nolint

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("unable to run command: %w: stderr: %s", err, string(stderrData))
	}

	return data, nil
}

func (c *toolsClipboardManager) Write(data []byte) error {

	cmd := exec.Command("wl-copy", "--trim-newline")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("unable to bind stdin: %w", err)
	}
	defer stdin.Close() //nolint

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("unable to bind stderr: %w", err)
	}
	defer stderr.Close() //nolint

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start command: %w", err)
	}

	if _, err := stdin.Write(data); err != nil {
		return fmt.Errorf("unable to write stdin: %w", err)
	}
	stdin.Close() //nolint

	stderrData, err := io.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("unable to read stderr: %w", err)
	}
	stderr.Close() // nolint

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("unable to run command: %w: stderr: %s", err, string(stderrData))
	}

	return nil
}

func (c *toolsClipboardManager) Watch(ctx context.Context) (<-chan []byte, <-chan error) {

	chout := make(chan []byte)
	cherr := make(chan error)

	go func() {

		for {
			cmd := exec.Command("wl-paste", "--no-newline", "-w", "echo")

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				cherr <- fmt.Errorf("unable to bind stdout: %w", err)
				break
			}
			defer stdout.Close() //nolint

			if err := cmd.Start(); err != nil {
				cherr <- fmt.Errorf("unable to start command: %w", err)
				break
			}

			scan := bufio.NewScanner(stdout)

			go func() {
				for scan.Scan() {
					data, err := c.Read()
					if err != nil {
						cherr <- fmt.Errorf("unable to scan stdout: %w", err)
						return
					}

					select {
					case chout <- data:
					case <-ctx.Done():
					default:
					}
				}
			}()
			if err := cmd.Wait(); err != nil {
				cherr <- fmt.Errorf("error while listening wl-paste (restarting): %w", err)
				time.Sleep(1 * time.Second)
			}
		}
	}()

	return chout, cherr
}
