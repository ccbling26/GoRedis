package wait

import (
	"sync"
	"time"
)

// 在 sync.WaitGroup 的基础上添加等待时间
type Wait struct {
	wg sync.WaitGroup
}

func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

func (w *Wait) Done() {
	w.wg.Done()
}

func (w *Wait) Wait() {
	w.wg.Wait()
}

// WaitWithTimeout 一直阻塞直到 WaitGroup 计数为 0 或超时。返回 true 如果超时
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan struct{}, 1)
	go func() {
		defer close(c)
		w.Wait()
		c <- struct{}{}
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}
