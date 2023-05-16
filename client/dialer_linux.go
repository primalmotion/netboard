//go:build linux

package client

import (
	"context"
	"net"
)

func netDialContextFunc(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		KeepAlive: -1,
	}
	return dialer.DialContext(ctx, network, addr)
}
