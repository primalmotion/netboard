package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.aporeto.io/wsc"
)

func makeSubscribeWSHandler(dispatch *dispatcher) func(http.ResponseWriter, *http.Request) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool { return true },
	}

	return func(w http.ResponseWriter, r *http.Request) {

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("unable to upgrade connection: %s", err),
				http.StatusBadRequest,
			)
			return
		}

		conn, err := wsc.Accept(r.Context(), ws, wsc.Config{
			PingPeriod: 15 * time.Minute,
			PongWait:   20 * time.Minute,
		})
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("unable to accept ws: %s", err),
				http.StatusBadRequest,
			)
			return
		}

		id := computeID(r)
		dispatch.Register(id)
		defer dispatch.Unregister(id)
		ch := dispatch.GetChannel(id)

		for {
			select {

			case data := <-ch:
				conn.Write(data)

			case <-conn.Done():
				return

			case <-r.Context().Done():
				conn.Close(websocket.CloseGoingAway)
			}
		}
	}
}
