package nats

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/quic-go/quic-go"
)

const quicScheme = "quic"

type quicConnStream struct {
	quic.Connection
	quic.Stream
}

func (c *quicConnStream) Close() error {
	return errors.Join(
		c.Stream.Close(),
		c.Connection.CloseWithError(0, "connection closed"),
	)
}

type quicDialer struct {
	tlsConfig  *tls.Config
	quicConfig *quic.Config
}

func (d *quicDialer) Dial(network, addr string) (net.Conn, error) {
	conn, err := quic.DialAddr(context.Background(), addr, d.tlsConfig, d.quicConfig)
	if err != nil {
		return nil, fmt.Errorf("quic.DialAddr: %w", err)
	}
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return nil, fmt.Errorf("conn.AcceptStream: %w", errors.Join(err, conn.CloseWithError(0, err.Error())))
	}
	return &quicConnStream{
		Connection: conn,
		Stream:     stream,
	}, nil
}

func isQUICScheme(u *url.URL) bool {
	return u.Scheme == quicScheme
}
