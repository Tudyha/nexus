package smux

import (
	"time"

	"github.com/xtaci/smux/v2"
)

// DefaultConfig 返回适用于隧道场景的 smux 默认配置，
// 在吞吐量和延迟之间取得平衡。
func DefaultConfig() *smux.Config {
	return &smux.Config{
		KeepAliveInterval: 10 * time.Second,
		KeepAliveTimeout:  30 * time.Second,
		MaxFrameSize:      65535,          // 64KB (smux 限制最大值 65535)
		MaxReceiveBuffer:  8 * 1024 * 1024, // 8MB — 提高并发流吞吐
		MaxStreamBuffer:   4 * 1024 * 1024, // 4MB — 提高单流缓冲
	}
}
