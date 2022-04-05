package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"ptplay"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("provide address to listen as the first argument")
	}

	l, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		return err
	}

	s := &http.Server{
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		Handler: ptplay.NewServer(ptplay.ServerOptions{
			DefaultHandler: func(conn ptplay.ConnectionWritable, bs []byte) {
				var request ptplay.Request
				if err := request.Unmarshal(bs); err != nil {
					log.Println(err)
					return
				}

				bs, err = useCaseRequest(request).Marshal()
				if err != nil {
					log.Println(err)
					return
				}

				if err := conn.Write(bs); err != nil {
					log.Println(err)
				}
			},
		}),
	}

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)

		sig := <-sigs
		log.Printf("terminate signal %v\n", sig)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			log.Println(err)
		}
	}()

	if err = s.Serve(l); err != nil {
		if err != http.ErrServerClosed {
			return err
		}

		log.Println("server stopped")
		return nil
	}

	return nil
}

func useCaseRequest(request ptplay.Request) ptplay.Response {
	return ptplay.Response{
		S: request.A + request.B,
		M: request.A * request.B,
	}
}
