package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func executeTask(task string) (string, error) {
	// 模拟任务执行
	fmt.Println("Executing task:", task)
	time.Sleep(2 * time.Second) // 模拟任务执行时间
	return "Task completed: " + task, nil
}

func main() {
	// 配置 Master 的 IP 地址
	masterIP := "127.0.0.1" // 替换为 Master 的 IP 地址
	masterPort := "9000"

	// 与 Master 建立 TCP 连接
	conn, err := net.Dial("tcp", masterIP+":"+masterPort)
	if err != nil {
		fmt.Println("Error connecting to master:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to master, waiting for tasks...")

	// 从 Master 接收任务
	task, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error receiving task from master:", err)
		return
	}
	task = strings.TrimSpace(task)

	// 执行任务
	result, err := executeTask(task)
	if err != nil {
		fmt.Println("Error executing task:", err)
		return
	}

	// 发送执行结果给 Master
	_, err = conn.Write([]byte(result + "\n"))
	if err != nil {
		fmt.Println("Error sending result to master:", err)
		return
	}

	fmt.Println("Task completed and result sent to master.")
}
