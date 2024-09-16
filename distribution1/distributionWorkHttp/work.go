package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type Task struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Body   string `json:"body,omitempty"`
}

func executeTask(task Task) (int, time.Duration, error) {
	start := time.Now()
	var statusCode int
	var err error
	if task.Method == "GET" {
		resp, err := http.Get(task.URL)
		if err == nil {
			statusCode = resp.StatusCode
		}
		resp.Body.Close()
	} else if task.Method == "POST" {
		resp, err := http.Post(task.URL, "application/json", bytes.NewBuffer([]byte(task.Body)))
		if err == nil {
			statusCode = resp.StatusCode
		}
		resp.Body.Close()
	}
	duration := time.Since(start)
	return statusCode, duration, err
}

func workerHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statusCode, duration, err := executeTask(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"status_code": statusCode,
		"duration":    duration.String(),
	}
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/run_task", workerHandler)
	http.ListenAndServe(":8080", nil)
	//http.ListenAndServe(":8081", nil)
	//http.ListenAndServe(":8082", nil)
}
