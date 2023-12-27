package tcp

import (
	"GoRedis/lib/logger"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect int           `yaml:"max_connect"`
	Timeout    time.Duration `yaml:"timeout"`
}

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

// ListenAndServe 监听并提供服务，并在收到 closeChan 发来的关闭通知后关闭
func ListenAndServe(listener net.Listener, handler Handler, closeChan <-chan struct{}) {
	// 监听关闭通知
	go func() {
		<-closeChan
		logger.Info("shutting down...")
		// 停止监听，listener.Accept() 会立即返回 io.EOF
		_ = listener.Close()
		// 关闭应用服务器
		_ = handler.Close()
	}()

	// 在异常退出后释放资源
	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		// 监听端口，阻塞直到收到新连接或出现错误
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		// 开启 goroutine 来处理新连接
		logger.Info("Aceept link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}

// ListenAndServeWithSignal 监听中断信号并通过 closeChan 通知服务器关闭
func ListenAndServeWithSignal(cfg *Config, handler Handler) error {
	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	// 如果出现列出的信号，传入到 sigChan 中
	// siganl 不会为了向 sigChan 发送信息而阻塞（即如果信号发送时如果 sigChan 阻塞了，signal 包会直接放弃）
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Bind %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}
