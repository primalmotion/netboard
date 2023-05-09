package client

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
)

func Copy(data io.Reader, url string, tlsConfig *tls.Config) error {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	r, err := http.NewRequest(http.MethodPost, url+"/copy", data)
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

	log.Println("data dispatched")

	return nil
}

func Listen(url string, tlsConfig *tls.Config) (chan []byte, error) {

	ch := make(chan []byte)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	r, err := http.NewRequest(http.MethodGet, url+"/paste", nil)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := client.Do(r)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("unable to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		close(ch)
		return nil, fmt.Errorf("server rejected the request: %w", err)
	}

	log.Println("connected and waiting for data")

	go func() {
		defer resp.Body.Close()

		buf := make([]byte, 1024)
		for {
			var chunk []byte
			for {
				n, err := resp.Body.Read(buf)
				if err != nil {
					close(ch)
					log.Printf("error: unable to read body: %s", err)
					return
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
				log.Printf("error: unable to decode body: %s", err)
				close(ch)
				return
			}

			ch <- decoded[:n]

			log.Println("data received")
		}
	}()

	return ch, nil
}
