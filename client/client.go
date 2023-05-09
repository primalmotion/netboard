package client

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func Copy(url string, tlsConfig *tls.Config) error {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	r, err := http.NewRequest(http.MethodPost, url+"/copy", os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server rejected the request: %s", resp.Status)
	}

	return nil
}

func Paste(url string, tlsConfig *tls.Config, command string, args ...string) error {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	r, err := http.NewRequest(http.MethodGet, url+"/paste", nil)
	if err != nil {
		return fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		var chunk []byte
		for {
			n, err := resp.Body.Read(buf)
			if err != nil {
				return fmt.Errorf("unable to read body: %w", err)
			}

			if buf[n-1] == ',' {
				chunk = append(chunk, buf[:n-1]...)
				break
			}
			chunk = append(chunk, buf[:n]...)
		}

		decoded := make([]byte, len(chunk))
		n, err := base64.RawURLEncoding.Decode(decoded, chunk)
		if err != nil {
			return fmt.Errorf("unable to decode data: %w", err)
		}

		cmd := exec.Command(command, args...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("unable to execute %s: %w", command, err)
		}

		cmd.Start()
		io.Copy(stdin, bytes.NewBuffer(decoded[:n]))
		stdin.Close()

		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("unable to run command %s: %w", command, err)
		}
	}
}
