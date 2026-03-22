package tools

import (
	"fmt"
	"runtime"
	"time"
)

// MemoryStats 内存统计信息
type MemoryStats struct {
	Alloc      uint64 // 当前分配的内存量
	TotalAlloc uint64 // 总共分配的内存量
	Sys        uint64 // 从系统获取的内存量
	NumGC      uint32 // GC 次数
}

// GetMemoryStats 获取当前内存使用情况
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemoryStats{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}
}

// MonitorMemory 监控内存使用情况，定期输出统计信息
func MonitorMemory(duration time.Duration) {
	startStats := GetMemoryStats()
	startTime := time.Now()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timer := time.NewTimer(duration)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			currentStats := GetMemoryStats()
			elapsed := time.Since(startTime)
			fmt.Printf("[Memory Monitor] Elapsed: %v | Alloc: %s | TotalAlloc: %s | Sys: %s | GC: %d\n",
				elapsed,
				SizeFormat(float64(currentStats.Alloc)),
				SizeFormat(float64(currentStats.TotalAlloc)),
				SizeFormat(float64(currentStats.Sys)),
				currentStats.NumGC-startStats.NumGC,
			)
		case <-timer.C:
			finalStats := GetMemoryStats()
			elapsed := time.Since(startTime)
			fmt.Printf("[Memory Monitor] Final - Elapsed: %v | Alloc: %s | TotalAlloc: %s | Sys: %s | GC: %d\n",
				elapsed,
				SizeFormat(float64(finalStats.Alloc)),
				SizeFormat(float64(finalStats.TotalAlloc)),
				SizeFormat(float64(finalStats.Sys)),
				finalStats.NumGC-startStats.NumGC,
			)
			return
		}
	}
}
