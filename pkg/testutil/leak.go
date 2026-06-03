package testutil

import (
	"runtime"
	"time"
)

// WaitForNumGoroutines 等待 goroutine 数量降到 target 以下，
// 返回最终 goroutine 数量。用于检测 goroutine 泄漏。
func WaitForNumGoroutines(target int, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	for {
		n := runtime.NumGoroutine()
		if n <= target {
			return n
		}
		if time.Now().After(deadline) {
			return n
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// CheckGoroutineLeak 返回比 target 多的 goroutine 数量。
// 在测试结束时调用，检测是否泄漏。
func CheckGoroutineLeak(target int) int {
	n := WaitForNumGoroutines(target, 5*time.Second)
	return n - target
}
