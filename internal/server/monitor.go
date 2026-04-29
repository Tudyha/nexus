package server

import (
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Resource struct {
	NumGoroutines int
	Alloc         uint64
	TotalAlloc    uint64
	Sys           uint64
	NumGC         uint32
}

type Monitor struct {
	stopCh   chan struct{}
	wg       sync.WaitGroup
	interval int32
}

func NewMonitor() Server {
	return &Monitor{
		stopCh:   make(chan struct{}),
		interval: 30,
	}
}

func (m *Monitor) Start() error {
	m.wg.Add(1)
	go m.monitor()

	log.Info().Msg("monitor started")
	return nil
}

func (m *Monitor) Stop() error {
	close(m.stopCh)
	m.wg.Wait()
	return nil
}

func (m *Monitor) monitor() {
	defer m.wg.Done()
	ticker := time.NewTicker(time.Duration(m.interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resource := m.collect()
			log.Printf("[Resource] goroutines=%d | Alloc=%d MB | TotalAlloc=%d MB | Sys=%d MB | NumGC=%d",
				resource.NumGoroutines, resource.Alloc, resource.TotalAlloc, resource.Sys, resource.NumGC)
		case <-m.stopCh:
			return
		}
	}
}

func (m *Monitor) collect() *Resource {
	// 1. Goroutine 数量
	numGoroutines := runtime.NumGoroutine()

	// 2. 内存统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 转换为 MB 便于阅读
	allocMB := memStats.Alloc / 1024 / 1024
	totalAllocMB := memStats.TotalAlloc / 1024 / 1024
	sysMB := memStats.Sys / 1024 / 1024

	return &Resource{
		NumGoroutines: numGoroutines,
		Alloc:         allocMB,
		TotalAlloc:    totalAllocMB,
		Sys:           sysMB,
		NumGC:         memStats.NumGC,
	}
}
