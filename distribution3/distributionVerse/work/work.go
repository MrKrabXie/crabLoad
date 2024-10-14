package main

import (
	"bufio"
	"crabloadtester/distribution3/common"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	concurrency int    // 存储并发量
	task        string // 存储任务
)

func executeTask(task string) (string, error) {
	// 模拟任务执行
	fmt.Println("Executing task:", task)
	time.Sleep(2 * time.Second) // 模拟任务执行时间
	return "Task completed: " + task, nil
}

// 执行任务
func handleWorkerTask(task string, concurrency int, waitStrategy common.WaitStrategy) {

	if concurrency <= 0 || task == "" {
		fmt.Println("No tasks to execute or concurrency is invalid.")
		return // 如果没有任务或并发量无效，则返回
	}

	var wg sync.WaitGroup
	taskCh := make(chan string, concurrency)

	// 创建并发的 worker goroutines 来执行任务
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				// 执行任务
				result, err := executeTask(task)
				if err != nil {
					fmt.Println("Error executing task:", err)
					return
				}
				fmt.Println(result)

				// 等待策略
				waitStrategy.Wait()
			}
		}()
	}

	// 将任务推送到 channel 中
	taskCh <- task
	close(taskCh)

	wg.Wait() // 等待所有并发任务完成
}

func main() {
	// 配置 Master 的 IP 地址
	masterIP := "127.0.0.1" // 替换为 Master 的 IP 地址
	masterPort := "9000"

	port := flag.String("port", masterPort, "port to listen on")
	if port != nil && *port != "" {
		masterPort = *port
	}
	flag.Parse()

	// 与 Master 建立 TCP 连接
	conn, err := net.Dial("tcp", masterIP+":"+masterPort)
	if err != nil {
		fmt.Println("Error connecting to master:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to master, waiting for tasks...")

	for {
		// 从 Master 接收任务
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error receiving task from master:", err)
			break
		}
		message = strings.TrimSpace(message)

		// 如果收到 NO_TASK 信号，表示暂时没有任务，等待一段时间
		if message == "NO_TASK" {
			fmt.Println("No new task from master. Waiting...")
			time.Sleep(5 * time.Second) // 等待5秒，再尝试接收任务
			continue
		}

		// 判断消息格式并提取任务和并发量
		if strings.HasPrefix(message, "task:") {
			task := strings.TrimPrefix(message, "task:")

			// 执行任务
			waitStrategy := &common.RandomWait{Min: 1 * time.Second, Max: 3 * time.Second}
			handleWorkerTask(task, concurrency, waitStrategy)

			// 发送执行结果给 master
			_, err = conn.Write([]byte("Task completed\n"))
			if err != nil {
				fmt.Println("Error sending result to master:", err)
				return
			}

			fmt.Println("Task completed and result sent to master.")

		} else if strings.HasPrefix(message, "concurrency:") {
			concurrencyStr := strings.TrimPrefix(message, "concurrency:")
			concurrencyStr = strings.TrimSpace(concurrencyStr) // 去掉前后空格

			var err error
			concurrency, err = strconv.Atoi(concurrencyStr) // 解析并发量
			if err != nil {
				fmt.Println("Error parsing concurrency from master:", err)
				continue
			}
			fmt.Printf("Concurrency updated to: %d\n", concurrency) // 打印更新后的并发量

		} else {
			fmt.Println("Invalid message format from master")
			continue
		}

		// 执行任务
		// 使用随机等待策略
		waitStrategy := &common.RandomWait{Min: 1 * time.Second, Max: 3 * time.Second}

		// 根据并发量执行任务
		handleWorkerTask(task, concurrency, waitStrategy)

		// 发送执行结果给 master
		_, err = conn.Write([]byte("Task completed\n"))
		if err != nil {
			fmt.Println("Error sending result to master:", err)
			return
		}

		fmt.Println("Task completed and result sent to master.")
	}

}
