package server

import (
	"log"
	"net/http"
)

func makeSubscribeChunkedHandler(dispatch *dispatcher) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusBadRequest)
			return
		}

		id := computeID(r)
		dispatch.Register(id)
		defer dispatch.Unregister(id)
		ch := dispatch.GetChannel(id)

		for {
			select {

			case <-r.Context().Done():
				flusher.Flush()
				return

			case c := <-ch:
				if _, err := w.Write(c); err != nil {
					log.Printf("unable to write chunk to client %s: %s", id, err)
				}
				flusher.Flush()
			}
		}
	}
}
