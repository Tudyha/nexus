package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket 令牌桶限流器（单机版，按 key 隔离）。
type TokenBucket struct {
	mu       sync.RWMutex
	buckets  map[string]*bucket
	rate     float64 // 每秒填充令牌数
	burst    int     // 最大令牌数
	stopCh   chan struct{}
	stopOnce sync.Once
}

type bucket struct {
	tokens   float64
	lastFill time.Time
}

// New 创建一个新的令牌桶限流器，rate 为每秒令牌数，burst 为最大突发。
func New(rate float64, burst int) *TokenBucket {
	tb := &TokenBucket{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   burst,
		stopCh:  make(chan struct{}),
	}
	// 每分钟清理一次过期 bucket
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tb.cleanup()
			case <-tb.stopCh:
				return
			}
		}
	}()
	return tb
}

// Allow 检查 key 是否允许通过本次请求。
func (tb *TokenBucket) Allow(key string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	b, ok := tb.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(tb.burst), lastFill: time.Now()}
		tb.buckets[key] = b
	}

	// 填充令牌
	now := time.Now()
	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens += elapsed * tb.rate
	if b.tokens > float64(tb.burst) {
		b.tokens = float64(tb.burst)
	}
	b.lastFill = now

	// 消费令牌
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// cleanup 清理超过 5 分钟未活跃的 bucket
func (tb *TokenBucket) cleanup() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	cutoff := time.Now().Add(-5 * time.Minute)
	for key, b := range tb.buckets {
		if b.lastFill.Before(cutoff) {
			delete(tb.buckets, key)
		}
	}
}

// Stop 停止后台清理协程。
func (tb *TokenBucket) Stop() {
	tb.stopOnce.Do(func() {
		close(tb.stopCh)
	})
}
