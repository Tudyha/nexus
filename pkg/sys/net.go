package sys

import (
	"time"

	"github.com/shirou/gopsutil/v4/net"
)

var (
	stats = make(map[time.Time]net.IOCountersStat)
	last  time.Time
)

func NetSpeed() (bytesSent float64, bytesRecv float64) {
	currStats, err := net.IOCounters(false)
	if err != nil || len(currStats) == 0 {
		return 0, 0
	}

	// 当前网络状态
	currTime := time.Now()
	currStat := currStats[0]
	stats[currTime] = currStat

	if last.IsZero() {
		last = currTime
		return 0, 0
	}

	prevStat, ok := stats[last]
	if !ok {
		last = currTime
		return 0, 0
	}
	timeDiff := currTime.Sub(last).Seconds()
	if timeDiff <= 0 {
		return 0, 0
	}

	bytesRecv = float64(currStat.BytesRecv-prevStat.BytesRecv) / timeDiff
	bytesSent = float64(currStat.BytesSent-prevStat.BytesSent) / timeDiff

	// Clean up old stats to prevent memory leak
	for t := range stats {
		if t.Before(currTime.Add(-5 * time.Minute)) {
			delete(stats, t)
		}
	}
	last = currTime
	return bytesSent, bytesRecv
}
