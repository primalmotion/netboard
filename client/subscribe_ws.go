package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.aporeto.io/wsc"
)

// SubscribeWS connects to the remote server and will get clipbiard updates using
// websockets.
func SubscribeWS(ctx context.Context, url string, tlsConfig *tls.Config) (chan []byte, chan struct{}) {

	ch := make(chan []byte, 512)
	done := make(chan struct{})

	go func() {

		isReconnect := false

	MAINWS:
		for {

			select {
			case <-ctx.Done():
				close(done)
				return
			default:
			}

			if isReconnect {
				time.Sleep(time.Second)
			}
			isReconnect = true

			conn, resp, err := wsc.Connect(
				ctx,
				strings.Replace(url+"/subscribe/ws", "https", "wss", 1),
				wsc.Config{TLSConfig: tlsConfig},
			)
			if err != nil {
				log.Printf("unable to connect to ws (retrying): %s", err)
				continue
			}

			if resp.StatusCode != http.StatusSwitchingProtocols {
				log.Printf("server rejected ws connection (retrying): %s", err)
				continue
			}

			log.Printf("connected to ws server")

			for {
				select {

				case msg := <-conn.Done():
					if !websocket.IsCloseError(msg, websocket.CloseGoingAway) {
						log.Printf("ws msg: %s", msg)
					} else {
						log.Println("ws server gone. reconnecting...")
					}
					continue MAINWS

				case data := <-conn.Read():

					data = bytes.TrimSuffix(data, []byte{','})
					decoded := make([]byte, len(data))

					n, err := base64.RawURLEncoding.Decode(decoded, data)
					if err != nil {
						log.Printf("error: unable to decode body: %s", err)
						continue
					}

					select {
					case ch <- decoded[:n]:
					default:
					}

				case <-ctx.Done():
					conn.Close(websocket.CloseGoingAway)
					<-conn.Done()
					close(done)
					return
				}
			}
		}
	}()

	return ch, done
}
