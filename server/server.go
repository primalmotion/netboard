package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"
)

// Serve starts the server that will handle and dispatch changes
// to of the clipboard.
func Serve(ctx context.Context, listenAddr string, tlsConf *tls.Config) error {

	server := http.Server{
		Addr:      listenAddr,
		TLSConfig: tlsConf,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	dispatch := newDispatcher()
	http.HandleFunc("/publish", makePublishHandler(dispatch))
	http.HandleFunc("/subscribe/chunked", makeSubscribeChunkedHandler(dispatch))
	http.HandleFunc("/subscribe/ws", makeSubscribeWSHandler(dispatch))

	// Start the server in a go routine
	srvErrCh := make(chan error)
	go func() {
		err := server.ListenAndServeTLS("", "")
		if !errors.Is(err, http.ErrServerClosed) {
			srvErrCh <- err
		}
	}()

	// Wait for a shutdown indicator to either return the
	// error or gracefully shutdown the server
	select {
	case err := <-srvErrCh:
		return err
	case <-ctx.Done():
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(closeCtx)
	}
}
