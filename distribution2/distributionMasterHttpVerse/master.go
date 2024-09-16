package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

func handleWorkerConnection(conn net.Conn, task string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	// 发送任务给 Worker
	_, err := conn.Write([]byte(task + "\n"))
	if err != nil {
		fmt.Println("Error sending task to worker:", err)
		return
	}

	// 接收 Worker 的执行结果
	result, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error receiving result from worker:", err)
		return
	}

	fmt.Println("Received result from worker:", strings.TrimSpace(result))
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:9000")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return
	}
	defer listener.Close()

	// 模拟一个任务
	task := "GET http://www.baidu.com"

	fmt.Println("Master node is running, waiting for workers to connect...")
	var wg sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// 每个 Worker 连接都在一个 goroutine 中处理
		wg.Add(1)
		go handleWorkerConnection(conn, task, &wg)
	}

	// 在这里你可能希望在关闭程序前等待所有任务完成
	// wg.Wait()
	// fmt.Println("All tasks completed.")
}
