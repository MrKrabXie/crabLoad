package common

import (
	"math/rand"
	"sync"
	"time"
)

// 定义等待策略接口
type WaitStrategy interface {
	Wait()
}

// RandomWait 实现等待策略
type RandomWait struct {
	Min time.Duration
	Max time.Duration
}

func (r *RandomWait) Wait() {
	waitDuration := time.Duration(rand.Int63n(int64(r.Max-r.Min))) + r.Min
	time.Sleep(waitDuration)
}

// 用户结构体
type User struct {
	WaitStrategy WaitStrategy // 等待策略
	Concurrency  int          // 并发数
}

// 用户执行任务
func (u *User) RunTasks(tasks []func()) {
	var wg sync.WaitGroup
	taskCh := make(chan func(), len(tasks))

	// 把所有任务放入 channel
	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	// 使用 goroutines 控制并发数
	for i := 0; i < u.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				u.WaitStrategy.Wait() // 等待策略
				task()                // 执行任务
			}
		}()
	}

	wg.Wait() // 等待所有 goroutines 完成
}

func NewUser(concurrency int, waitStrategy WaitStrategy) *User {
	return &User{
		WaitStrategy: waitStrategy,
		Concurrency:  concurrency,
	}
}
