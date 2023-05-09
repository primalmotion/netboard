package server

import (
	"sync"
)

type Dispatcher interface {
	Register(string)
	Unregister(string)
	Dispatch(string, []byte)
	GetChannel(string) chan []byte
}

type dispatcher struct {
	sync.RWMutex
	clients map[string]chan []byte
}

func NewDispatcher() Dispatcher {
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
