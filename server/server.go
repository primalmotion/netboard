package server

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Serve starts the server that will handle and dispatch changes
// to of the clipboard.
func Serve(listenAddr string, tlsConf *tls.Config) error {
	dispatch := newDispatcher()

	server := http.Server{
		Addr:      listenAddr,
		TLSConfig: tlsConf,
	}

	http.HandleFunc("/copy", func(w http.ResponseWriter, r *http.Request) {

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to read body: %s", err), http.StatusBadRequest)
		}

		encoded := base64.RawURLEncoding.EncodeToString(data)
		encoded = encoded + ","

		id := computeID(r)
		log.Printf("dispatched data from: %s", id)

		dispatch.Dispatch(id, []byte(encoded))
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/paste", func(w http.ResponseWriter, r *http.Request) {

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusBadRequest)
			return
		}

		id := computeID(r)

		dispatch.Register(id)
		log.Printf("client registered: %s", id)
		defer func() {
			dispatch.Unregister(id)
			log.Printf("client unregistered: %s", id)
		}()

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
	})

	return server.ListenAndServeTLS("", "")
}

func computeID(r *http.Request) string {
	cert := r.TLS.PeerCertificates[0]
	return fmt.Sprintf("%02X", sha256.Sum256(cert.Raw)) // #nosec
}
