package ptplay

import (
	"context"
	"errors"

	"nhooyr.io/websocket"
)

type Connection interface {
	ConnectionReadable
	ConnectionWritable
}

type ConnectionWritable interface {
	Write(bs []byte) error
}

type ConnectionReadable interface {
	Read() ([]byte, error)
}

type connection struct {
	c   *websocket.Conn
	ctx context.Context
}

func (c *connection) Write(bs []byte) error {
	return c.c.Write(context.Background(), websocket.MessageText, bs)
}

func (c *connection) Read() ([]byte, error) {
	for {
		select {
		case <-context.Background().Done():
			return nil, context.Background().Err()
		default:

		}

		t, m, err := c.c.Read(context.Background())
		if err != nil {
			return nil, err
		}

		if t == websocket.MessageBinary {
			return nil, errors.New("read only text message type ")
		}

		return m, nil
	}
}
