package server

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync"
)

func computeID(r *http.Request) string {
	cert := r.TLS.PeerCertificates[0]
	return fmt.Sprintf("%02X", sha256.Sum256(cert.Raw)) // #nosec
}

type dispatcher struct {
	sync.RWMutex
	clients map[string]chan []byte
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		clients: make(map[string]chan []byte),
	}
}

func (d *dispatcher) Register(c string) {
	d.Lock()
	defer d.Unlock()

	d.clients[c] = make(chan []byte)
}

func (d *dispatcher) Unregister(c string) {
	d.Lock()
	defer d.Unlock()

	if _, ok := d.clients[c]; !ok {
		return
	}

	close(d.clients[c])
	delete(d.clients, c)
}

func (d *dispatcher) Dispatch(srcID string, data []byte) {
	d.RLock()
	defer d.RUnlock()

	for id, c := range d.clients {
		if srcID == id {
			continue
		}
		select {
		case c <- data:
		default:
		}
	}
}

func (d *dispatcher) GetChannel(c string) chan []byte {

	d.RLock()
	defer d.RUnlock()

	return d.clients[c]
}
