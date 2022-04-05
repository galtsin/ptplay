package ptplay

import (
	"context"
	"log"
	"net/http"

	"nhooyr.io/websocket"
)

type ServerOptions struct {
	DefaultHandler HandlerFunc
}

type HandlerFunc func(ConnectionWritable, []byte)

type Server struct {
	opts ServerOptions
}

func NewServer(opts ServerOptions) *Server {
	return &Server{
		opts: opts,
	}
}

// Новое соединение
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer c.Close(websocket.StatusInternalError, "")

	log.Println("connection was created")

	if err := s.readLoop(r.Context(), &connection{c: c}); err != nil {
		log.Println(err)
	}
}

func (s *Server) readLoop(ctx context.Context, conn Connection) error {
	for {
		select {
		case <-ctx.Done():
			break
		default:
			bs, err := conn.Read()
			if err != nil {
				return err
			}

			if s.opts.DefaultHandler != nil {
				s.opts.DefaultHandler(conn, bs)
			}
		}
	}
}
