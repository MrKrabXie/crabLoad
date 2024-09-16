package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// Task 接口，定义任务执行的通用行为
type Task interface {
	Execute() error
}

// GetRequestTask 实现 GET 请求任务
type GetRequestTask struct {
	URL string
}

func (t *GetRequestTask) Execute() error {
	startTime := time.Now()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(t.URL)

	err := fasthttp.Do(req, resp)
	if err != nil {
		return err
	}
	duration := time.Since(startTime)
	//fmt.Println("GET %s - Status Code: %d - Duration: %v", t.URL, resp.StatusCode(), duration)
	fmt.Printf("GET %s - Status Code: %d - Duration: %v \n", t.URL, resp.StatusCode(), duration)
	recordResult(resp.StatusCode(), duration)
	return nil
}

// PostRequestTask 实现 POST 请求任务
type PostRequestTask struct {
	URL  string
	Body []byte
}

func (t *PostRequestTask) Execute() error {
	startTime := time.Now()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(t.URL)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody(t.Body)

	err := fasthttp.Do(req, resp)
	if err != nil {
		return err
	}
	duration := time.Since(startTime)
	fmt.Printf("POST %s - Status Code: %d - Duration: %v\\n", t.URL, resp.StatusCode(), duration)
	recordResult(resp.StatusCode(), duration)
	return nil
}

// User 定义了用户的行为
type User struct {
	Tasks        []Task
	WaitStrategy WaitStrategy
}

func (u *User) Run() {
	for _, task := range u.Tasks {
		err := task.Execute()
		if err != nil {
			fmt.Println("Error executing task:", err)
		}
		time.Sleep(u.WaitStrategy.Wait()) // 等待下一个任务执行
	}
}

// WaitStrategy 接口定义了等待策略
type WaitStrategy interface {
	Wait() time.Duration
}

// ConstantWait 是一个固定等待时间的策略
type ConstantWait struct {
	Duration time.Duration
}

func (cw *ConstantWait) Wait() time.Duration {
	return cw.Duration
}

// RandomWait 是一个随机等待时间的策略
type RandomWait struct {
	Min time.Duration
	Max time.Duration
}

func (rw *RandomWait) Wait() time.Duration {
	return time.Duration(rand.Int63n(int64(rw.Max-rw.Min))) + rw.Min
}

// 结果收集
var results = struct {
	sync.Mutex
	StatusCodes map[int]int
	Latencies   []time.Duration
}{
	StatusCodes: make(map[int]int),
	Latencies:   []time.Duration{},
}

func recordResult(statusCode int, latency time.Duration) {
	results.Lock()
	defer results.Unlock()
	results.StatusCodes[statusCode]++
	results.Latencies = append(results.Latencies, latency)
}

// generateReport 输出测试报告
func generateReport() {
	results.Lock()
	defer results.Unlock()

	fmt.Println("Status Codes:", results.StatusCodes)
	fmt.Println()

	var totalLatency time.Duration
	for _, latency := range results.Latencies {
		totalLatency += latency
	}
	avgLatency := totalLatency / time.Duration(len(results.Latencies))
	fmt.Println("Average Latency:", avgLatency)
	fmt.Println()
}

// LoadTester 负责启动并发用户模拟负载
type LoadTester struct {
	NumUsers int
	Duration time.Duration
}

func (lt *LoadTester) Start() {
	var wg sync.WaitGroup
	for i := 0; i < lt.NumUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			user := User{
				Tasks: []Task{
					&GetRequestTask{URL: "https://www.baidu.com"},
					//&PostRequestTask{URL: "https://www.baidu.com"},
					//&PostRequestTask{URL: "https://www.baidu.com", Body: []byte(`{"key": "value"}`)},
				},
				WaitStrategy: &RandomWait{Min: time.Second, Max: 2 * time.Second},
			}
			startTime := time.Now()
			for time.Since(startTime) < lt.Duration {
				user.Run()
			}
		}(i)
	}
	wg.Wait()
}

func main() {
	tester := LoadTester{
		NumUsers: 10,
		Duration: 10 * time.Second,
	}
	tester.Start()

	// 输出结果报告
	generateReport()
}
