package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"log"
	"net/http"
	"time"
)

// SubscribeChunked connects to the remote server and will get clipbiard updates using
// HTTP chunked encoding.
func SubscribeChunked(ctx context.Context, url string, tlsConfig *tls.Config) (chan []byte, chan struct{}) {

	ch := make(chan []byte, 512)
	done := make(chan struct{})

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

			r, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/subscribe/chunked", nil)
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
			defer resp.Body.Close() // nolint

			for {
				var chunk []byte
				for {
					buf := make([]byte, 1024)
					n, err := resp.Body.Read(buf)
					if err != nil {
						log.Printf("error: unable to read body: %s", err)
						continue MAIN
					}

					if n == 0 {
						break
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
					log.Println("data received: sent to channel")
				case <-ctx.Done():
					close(done)
					return
				default:
					log.Println("data received: channel full")
				}
			}
		}
	}()

	return ch, done
}
