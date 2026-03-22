package tools

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 模拟高并发请求的测试函数
func simulateRequest() {
	// 模拟一些常见的操作
	Base64encode("test message for base64 encoding")
	RandString(20)
	Md5("test string for md5 hashing")
}

// BenchmarkHighConcurrency 基准测试高并发场景
func BenchmarkHighConcurrency(b *testing.B) {
	// 重置计时器
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		concurrency := 100 // 并发数

		// 启动并发请求
		for j := 0; j < concurrency; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				simulateRequest()
			}()
		}

		// 等待所有请求完成
		wg.Wait()
	}
}

// TestHighConcurrencyWithMemoryMonitor 测试高并发场景下的内存使用情况
func TestHighConcurrencyWithMemoryMonitor(t *testing.T) {
	// 启动内存监控
	go MonitorMemory(10 * time.Second)

	// 模拟1000/s的请求，持续10秒
	requestsPerSecond := 1000
	duration := 10 * time.Second
	totalRequests := requestsPerSecond * int(duration.Seconds())

	var wg sync.WaitGroup
	wg.Add(totalRequests)

	// 控制请求速率
	ticker := time.NewTicker(time.Second / time.Duration(requestsPerSecond))
	defer ticker.Stop()

	startTime := time.Now()
	requestCount := 0

	for time.Since(startTime) < duration {
		<-ticker.C
		if requestCount < totalRequests {
			go func() {
				defer wg.Done()
				simulateRequest()
			}()
			requestCount++
		}
	}

	// 等待所有请求完成
	wg.Wait()

	fmt.Printf("Test completed: %d requests in %v\n", requestCount, duration)
}
