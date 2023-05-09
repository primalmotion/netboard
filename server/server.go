package server

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

func Serve(listenAddr string, tlsConf *tls.Config) error {
	dispatch := NewDispatcher()

	server := http.Server{
		Addr:      listenAddr,
		TLSConfig: tlsConf,
	}

	http.HandleFunc("/copy", func(w http.ResponseWriter, r *http.Request) {

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to read body: %s", err), http.StatusBadRequest)
		}
		defer r.Body.Close()

		encoded := base64.RawURLEncoding.EncodeToString(data)
		encoded = encoded + ","

		dispatch.Dispatch([]byte(encoded))
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/paste", func(w http.ResponseWriter, r *http.Request) {

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusBadRequest)
			return
		}

		dispatch.Register(r.RemoteAddr)
		defer dispatch.Unregister(r.RemoteAddr)

		ch := dispatch.GetChannel(r.RemoteAddr)

		for {
			select {
			case <-r.Context().Done():
				flusher.Flush()
				return
			case c := <-ch:
				w.Write(c)
				flusher.Flush()
			}
		}
	})

	return server.ListenAndServeTLS("", "")
}
