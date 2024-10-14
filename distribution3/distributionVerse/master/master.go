package main

import (
	"bufio"
	"crabloadtester/distribution3/common"    // 引入 common 包
	"crabloadtester/distribution3/scheduler" // 引入调度器包
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func handleWorkerConnection(conn net.Conn, user *common.User, sched *scheduler.Scheduler, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	tasks := []string{
		//"GET http://www.google.com",
		"GET http://www.baidu.com",
		//"GET http://www.example.com",
	}

	taskIndex := 0
	for {

		// 如果有剩余的任务，发送任务
		if taskIndex < len(tasks) {
			task := tasks[taskIndex]
			taskIndex++

			// 发送任务给 Worker
			message := fmt.Sprintf("task:%s \n", task) // 任务
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println("Error sending task to worker:", err)
				return
			}
		} else {
			// 没有任务时，发送心跳包 NO_TASK
			_, err := conn.Write([]byte("NO_TASK\n"))
			if err != nil {
				fmt.Println("Error sending NO_TASK signal:", err)
				break
			}

			// 等待一段时间再检查是否有新任务
			time.Sleep(5 * time.Second) // 5秒的心跳等待，可以根据需要调整
		}

		// 接收 Worker 的执行结果
		result, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error receiving result from worker:", err)
			break
		}

		fmt.Println("Received result from worker:", strings.TrimSpace(result))
	}

}

func main() {
	// 创建用户实例，设置并发数和等待策略
	concurrency := 12 // 可以从外部传入
	waitStrategy := &common.RandomWait{Min: 1 * time.Second, Max: 3 * time.Second}
	user := common.NewUser(concurrency, waitStrategy)

	sched := scheduler.NewScheduler() //生成调度器

	listener, err := net.Listen("tcp", "0.0.0.0:9000")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Master node is running, waiting for workers to connect...")
	var wg sync.WaitGroup
	// 创建一个信号通道
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}

			// 打印接受到的连接信息
			fmt.Printf("Accepted connection from worker: %s\n", conn.RemoteAddr().String())

			// 增加工作节点并打印分配信息
			sched.AddWorker(conn, user.Concurrency)               // 使用 user.Concurrency
			sched.PrintWorkerInfo()                               // 打印当前工作节点的并发信息
			fmt.Printf("Total Workers: %d\n", len(sched.Workers)) // Log the total number of workers

			// 每个 Worker 连接都在一个 goroutine 中处理
			wg.Add(1)
			go handleWorkerConnection(conn, user, sched, &wg)
		}
	}()

	// 等待系统信号以关闭程序
	<-signalChan // 阻塞，直到接收到信号
	fmt.Println("Shutting down the server...")

	// 关闭监听器
	listener.Close()

	// 等待所有任务完成
	wg.Wait()
	fmt.Println("All tasks completed.")
}
