package tcp

import (
	"GoRedis/lib/logger"
	"GoRedis/lib/sync/atomic"
	"GoRedis/lib/sync/wait"
	"bufio"
	"context"
	"io"
	"net"
	"sync"
	"time"
)

// Client 客户端连接
type Client struct {
	Conn    net.Conn  // tcp 连接
	Waiting wait.Wait // 当服务端开始发送数据时进入等待, 防止其它 goroutine 关闭连接
}

type EchoHandler struct {
	activeConn sync.Map          // 保存所有工作状态 client 的集合，需使用并发安全的容器
	closing    atomic.AtomicBool // 关闭状态标识位
}

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	// 关闭中的 handler 不会处理新的连接
	if h.closing.Get() {
		conn.Close()
		return
	}

	client := &Client{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("Connection closed")
				h.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}

		client.Waiting.Add(1) // 发送数据前先置为等待状态，阻止连接被关闭

		conn.Write([]byte(msg))

		client.Waiting.Done()
	}
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second) // 等待数据发送完成或超时
	c.Conn.Close()
	return nil
}

func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// 逐个关闭连接
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*Client)
		client.Close()
		return true
	})
	return nil
}
