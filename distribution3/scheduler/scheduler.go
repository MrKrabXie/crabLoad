package scheduler

import (
	"fmt"
	"net"
	"sync"
)

// Scheduler 结构体
type Scheduler struct {
	mu      sync.Mutex   // 用于保护 Workers 访问
	Workers []WorkerInfo // 存储工作节点的信息
}

// WorkerInfo 结构体，存储工作节点的连接和分配的并发数
type WorkerInfo struct {
	Conn        net.Conn
	Concurrency int
}

// NewScheduler 创建一个新的调度器
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// AddWorker 添加工作节点
func (s *Scheduler) AddWorker(conn net.Conn, totalConcurrency int) {
	s.mu.Lock()         // Lock the mutex
	defer s.mu.Unlock() // Ensure the mutex is unlocked when done

	// 先将 worker 加入列表
	s.Workers = append(s.Workers, WorkerInfo{Conn: conn, Concurrency: 0})

	// 进行并发量分配
	s.DistributeConcurrency(totalConcurrency)

	// 发送分配的并发数给工作节点
	message := fmt.Sprintf("concurrency: %d\n", s.Workers[len(s.Workers)-1].Concurrency)
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending concurrency to worker:", err)
	}
}

// DistributeConcurrency 根据工作节点数量分配并发量
func (s *Scheduler) DistributeConcurrency(totalConcurrency int) {
	numWorkers := len(s.Workers)
	if numWorkers == 0 {
		return
	}

	// 计算每个工作节点的并发量
	baseConcurrency := totalConcurrency / numWorkers
	remainder := totalConcurrency % numWorkers

	for i := range s.Workers {
		// 基本并发量
		s.Workers[i].Concurrency = baseConcurrency

		// 将剩余的并发量分配给前面几个工作节点
		if i < remainder {
			s.Workers[i].Concurrency++
		}
	}
}

// PrintWorkerInfo 打印工作节点的信息
func (s *Scheduler) PrintWorkerInfo() {
	for i, worker := range s.Workers {
		fmt.Printf("Worker %d: Concurrency = %d\n", i+1, worker.Concurrency)
	}
}
