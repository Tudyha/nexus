package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestAllow_Basic(t *testing.T) {
	tb := New(100, 100)
	defer tb.Stop()

	if !tb.Allow("key1") {
		t.Error("expected Allow to return true for first request")
	}
}

func TestAllow_ExhaustBurst(t *testing.T) {
	tb := New(1000, 5)
	defer tb.Stop()

	for i := range 5 {
		if !tb.Allow("key") {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	// 6th request within the same burst should be rate-limited
	if tb.Allow("key") {
		t.Error("request 6 should be rate-limited (burst=5)")
	}
}

func TestAllow_Refill(t *testing.T) {
	tb := New(1000, 1)
	defer tb.Stop()

	if !tb.Allow("key") {
		t.Fatal("first request should be allowed")
	}

	if tb.Allow("key") {
		t.Fatal("second immediate request should be rate-limited")
	}

	// wait for refill (at 1000 tokens/s, 1 token takes 1ms)
	time.Sleep(5 * time.Millisecond)

	if !tb.Allow("key") {
		t.Error("request after refill should be allowed")
	}
}

func TestAllow_KeyIsolation(t *testing.T) {
	tb := New(1000, 1)
	defer tb.Stop()

	if !tb.Allow("key-a") {
		t.Fatal("key-a first should be allowed")
	}

	if !tb.Allow("key-b") {
		t.Fatal("key-b first should be allowed (isolated from key-a)")
	}
}

func TestAllow_ZeroRate(t *testing.T) {
	tb := New(0, 0)
	defer tb.Stop()

	if tb.Allow("key") {
		t.Error("with rate=0, Allow should never return true (but should not panic)")
	}
}

func TestConcurrentAccess(t *testing.T) {
	tb := New(1000, 100)
	defer tb.Stop()

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				tb.Allow("shared-key")
			}
		}()
	}
	wg.Wait()

	// no panics = pass
}

func TestCleanup(t *testing.T) {
	tb := New(1000, 5)
	defer tb.Stop()

	tb.Allow("stale-key")

	if _, ok := tb.buckets["stale-key"]; !ok {
		t.Fatal("bucket should exist before cleanup")
	}

	// force cleanup by backdating the bucket
	tb.mu.Lock()
	tb.buckets["stale-key"].lastFill = time.Now().Add(-10 * time.Minute)
	tb.mu.Unlock()

	tb.cleanup()

	tb.mu.RLock()
	_, ok := tb.buckets["stale-key"]
	tb.mu.RUnlock()
	if ok {
		t.Error("bucket should have been cleaned up")
	}
}

func TestStop(t *testing.T) {
	tb := New(10, 5)

	tb.Stop()
	tb.Stop() // double-stop should not panic
}

func TestAllow_HighRate(t *testing.T) {
	tb := New(10000, 1000)
	defer tb.Stop()

	for range 1000 {
		if !tb.Allow("key") {
			t.Fatal("all 1000 requests within burst should be allowed")
		}
	}
}

func TestTokenBucket_RefillRate(t *testing.T) {
	tb := New(100, 100) // 100 tokens/s, burst 100
	defer tb.Stop()

	// exhaust burst
	for range 100 {
		tb.Allow("key")
	}

	// wait 500ms → ~50 tokens refilled
	time.Sleep(500 * time.Millisecond)

	allowed := 0
	for range 100 {
		if tb.Allow("key") {
			allowed++
		}
	}

	// should have ~50 tokens
	if allowed < 30 || allowed > 70 {
		t.Errorf("expected ~50 allowed tokens after 500ms, got %d", allowed)
	}
}
