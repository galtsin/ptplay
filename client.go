package ptplay

import (
	"context"
	"fmt"
	"net"
	"time"

	"nhooyr.io/websocket"
)

type ClientInterface interface {
	Connection() (Connection, error)
}

type ClientOptions struct {
	Addr net.Addr
}

type Client struct {
	opts ClientOptions
}

// Фабрика для создания соединений
func NewClient(opts ClientOptions) *Client {
	return &Client{opts: opts}
}

func (cl *Client) Connection() (Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s", cl.opts.Addr.String()), nil)
	if err != nil {
		return nil, err
	}

	return &connection{c: c}, nil
}
