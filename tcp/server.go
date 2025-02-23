package tcp

import (
	"context"
	"go-redis/interface/tcp"
	"go-redis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Addr string
}

// ListenAndServeWithSignal starts a tcp server
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	closeChan := make(chan struct{})
	// when the system signal is received, close the listener
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	logger.Info("Starting server on", cfg.Addr)
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe accepts connections on the Listener
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// User manually close the server
	go func() {
		// block until closeChan is signaled
		<-closeChan
		close(listener, handler)
	}()

	defer close(listener, handler)
	ctx := context.Background()
	// use wait group to wait all go routine to exit
	waitDone := sync.WaitGroup{}
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("Accepted connection from", conn.RemoteAddr().String())
		waitDone.Add(1)
		go func() {
			// use defer to prevent wait group not executed if goroutine panic
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}

func close(listener net.Listener, handler tcp.Handler) {
	_ = listener.Close()
	_ = handler.Close()
}
