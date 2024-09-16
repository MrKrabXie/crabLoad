package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Task struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Body   string `json:"body,omitempty"`
}

func sendTaskToWorker(workerURL string, task Task) {
	taskBytes, _ := json.Marshal(task)
	resp, err := http.Post(workerURL+"/run_task", "application/json", bytes.NewBuffer(taskBytes))
	if err != nil {
		fmt.Println("Error sending task to worker:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error decoding worker response:", err)
		return
	}

	fmt.Println("Worker result:", result)
}

func main() {
	workers := []string{
		"http://localhost:8080", // Worker 1
		"http://localhost:8081", // Worker 1
		"http://localhost:8082", // Worker 2
	}

	tasks := []Task{
		{Method: "GET", URL: "https://www.baidu.com"},
		{Method: "POST", URL: "https://www.baidu.com", Body: `{"key": "value"}`},
	}

	var wg sync.WaitGroup
	for _, worker := range workers {
		for _, task := range tasks {
			wg.Add(1)
			go func(worker string, task Task) {
				defer wg.Done()
				sendTaskToWorker(worker, task)
			}(worker, task)
		}
	}
	wg.Wait()
	fmt.Println("All tasks complete.")
}
