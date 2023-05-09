package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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

func Listen(ctx context.Context, url string, tlsConfig *tls.Config) chan []byte {

	ch := make(chan []byte, 512)

	go func() {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}

		isReconnect := false

	MAIN:
		for {

			if isReconnect {
				time.Sleep(5 * time.Second)
			}
			isReconnect = true

			r, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/paste", nil)
			if err != nil {
				log.Printf("unable to build request: %s", err)
				continue
			}

			log.Println("connected and waiting for data")

			resp, err := client.Do(r)
			if err != nil {
				log.Printf("unable to send request: %s", err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Printf("server rejected the request: %s", err)
				continue
			}
			defer resp.Body.Close()

			buf := make([]byte, 1024)

			for {
				var chunk []byte
				for {
					n, err := resp.Body.Read(buf)
					if err != nil {
						log.Printf("error: unable to read body: %s", err)
						continue MAIN
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
					continue
				}

				select {
				case ch <- decoded[:n]:
					log.Println("data received", string(decoded[:n]))
				case <-ctx.Done():
					return
				default:
					log.Println("unable to process received data: channel full")
				}

			}
		}
	}()

	return ch
}
