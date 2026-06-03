package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/Tudyha/nexus/pkg/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	globalLimiter *ratelimit.TokenBucket
	limiterOnce   sync.Once
)

// InitRateLimiter 初始化全局限流器。
// rate: 每秒令牌数，burst: 最大突发。rate<=0 表示不限流。
func InitRateLimiter(rate float64, burst int) {
	if rate <= 0 {
		return
	}
	limiterOnce.Do(func() {
		globalLimiter = ratelimit.New(rate, burst)
		log.Info().Float64("rate", rate).Int("burst", burst).Msg("rate limiter initialized")
	})
}

// RateLimit 返回基于客户端 IP 的速率限制中间件。
func RateLimit() gin.HandlerFunc {
	if globalLimiter == nil {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}
	return func(ctx *gin.Context) {
		ip := clientIP(ctx)
		if !globalLimiter.Allow(ip) {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
			return
		}
		ctx.Next()
	}
}

// clientIP 从请求中提取客户端 IP，优先考虑 X-Forwarded-For。
func clientIP(ctx *gin.Context) string {
	ip := ctx.Request.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}
	ip = ctx.Request.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	return ctx.ClientIP()
}
