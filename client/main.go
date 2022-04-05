package main

import (
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
}

func DefaultConfig() Config {
	return Config{
		ConnectionSize: 1,
		Interval:       500 * time.Millisecond,
	}
}

func main() {
	config := DefaultConfig()

	hFlag := flag.String("h", "", "server address")
	nFlag := flag.Int("n", config.ConnectionSize, "connections number")
	iFlag := flag.Float64("i", config.Interval.Seconds(), "interval, msec")

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
	wg := sync.WaitGroup{}

	for i := 0; i < config.ConnectionSize; i++ {
		if err := createWorker(client, transferChannel, &wg, i); err != nil {
			return err
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		ticker := time.NewTicker(config.Interval)
		for {
			select {
			case sig := <-sigs:
				log.Printf("terminating by %v\n", sig)
				close(transferChannel)
				return
			case <-ticker.C:
				transferChannel <- useCaseRequest()
			}
		}
	}()

	wg.Wait()
	return nil
}

func createWorker(client ptplay.ClientInterface, ch chan ptplay.Request, wg *sync.WaitGroup, index int) error {
	conn, err := client.Connection()
	if err != nil {
		return err
	}

	go func() {
		wg.Add(1)
		defer wg.Done()

		logger := log.New(os.Stdout, fmt.Sprintf("[w%d]", index), log.LstdFlags)

		for request := range ch {
			bs, err := request.Marshal()
			if err != nil {
				logger.Println(err)
				continue
			}

			if err := conn.Write(bs); err != nil {
				logger.Println(err)
				continue
			}

			bs, err = conn.Read()
			if err != nil {
				logger.Println(err)
				continue
			}

			var response ptplay.Response
			if err := response.Unmarshal(bs); err != nil {
				logger.Println(err)
				continue
			}

			bs, err = useCaseResult(request, response).Marshal()
			if err != nil {
				logger.Println(err)
				continue
			}

			// Вывод результата в stdout
			logger.Println(string(bs))
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

func randInt() int {
	x, _ := rand.Int(rand.Reader, big.NewInt(int64(1000)))
	return int(x.Int64())
}
