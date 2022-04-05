package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"ptplay"
)

type Config struct {
	ServerAddr     net.Addr
	ConnectionSize int
	Interval       time.Duration
	DefaultFile    string
}

func DefaultConfig() Config {
	return Config{
		ConnectionSize: 1,
		Interval:       1000 * time.Millisecond,
		DefaultFile:    "output.txt",
	}
}

func main() {
	config := DefaultConfig()

	hFlag := flag.String("h", "", "server address")
	nFlag := flag.Int("n", config.ConnectionSize, "connections number")
	iFlag := flag.Int64("i", config.Interval.Milliseconds(), "interval, msec")

	flag.Parse()

	addr, err := net.ResolveTCPAddr("tcp", *hFlag)
	if err != nil {
		panic(err)
	}

	if 0 >= *nFlag {
		panic("count should be more 0")
	}

	config.ServerAddr = addr
	config.Interval = time.Duration(*iFlag) * time.Millisecond
	config.ConnectionSize = *nFlag

	if err := run(config); err != nil {
		panic(err)
	}
}

func run(config Config) error {
	client := ptplay.NewClient(ptplay.ClientOptions{
		Addr: config.ServerAddr,
	})

	transferChannel := make(chan ptplay.Request, config.ConnectionSize)
	outputChannel := make(chan []byte, 0)

	wg := sync.WaitGroup{}

	mainContext, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	file, err := os.OpenFile(config.DefaultFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer log.Println("saving stopped")
		defer close(outputChannel)
		defer file.Close()

		for {
			select {
			case <-mainContext.Done():
				return
			case bs := <-outputChannel:
				if _, err := file.Write(bs); err != nil {
					log.Println(err)
				}
			}
		}
	}()

	for i := 0; i < config.ConnectionSize; i++ {
		if err := createWorker(client, transferChannel, &wg, i, outputChannel); err != nil {
			return err
		}
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer log.Println("transfer stopped")
		defer close(transferChannel)

		ticker := time.NewTicker(config.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-mainContext.Done():
				return
			case <-ticker.C:
				transferChannel <- useCaseRequest()
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	// Горутина, которая подчищает ресурсы
	go func() {
		sig := <-sigs
		log.Printf("terminating by %v\n", sig)

		mainCancel()
	}()

	wg.Wait()
	return nil
}

func createWorker(client ptplay.ClientInterface, ch chan ptplay.Request, wg *sync.WaitGroup, index int, out chan []byte) error {
	conn, err := client.Connection()
	if err != nil {
		return err
	}

	go func() {
		wg.Add(1)
		defer wg.Done()

		logger := log.New(os.Stdout, fmt.Sprintf("[w%d]", index), log.LstdFlags)

		for request := range ch {
			response, err := transfer(conn, request)
			if err != nil {
				logger.Println(err)
				continue
			}

			bs, err := useCaseResult(request, response).Marshal()
			if err != nil {
				logger.Println(err)
				continue
			}

			// DEBUG: Вывод результата в stdout
			logger.Println(string(bs))
			out <- bs
		}
	}()

	return nil
}

func useCaseRequest() ptplay.Request {
	return ptplay.Request{
		A: randInt(),
		B: randInt(),
	}
}

func useCaseResult(request ptplay.Request, response ptplay.Response) ptplay.Result {
	return ptplay.Result{
		Request:  request,
		Response: response,
	}
}

func transfer(conn ptplay.Connection, request ptplay.Request) (ptplay.Response, error) {
	var response ptplay.Response

	bs, err := request.Marshal()
	if err != nil {
		return response, err
	}

	if err := conn.Write(bs); err != nil {
		return response, err
	}

	bs, err = conn.Read()
	if err != nil {
		return response, err
	}

	if err := response.Unmarshal(bs); err != nil {
		return response, err
	}

	return response, nil
}

func randInt() int {
	x, _ := rand.Int(rand.Reader, big.NewInt(int64(1000)))
	return int(x.Int64())
}
