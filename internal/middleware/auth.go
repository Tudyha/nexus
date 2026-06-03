package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/Tudyha/nexus/internal/config"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/internal/service"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/appleboy/gin-jwt/v3/core"
	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

var (
	ginJwt           *jwt.GinJWTMiddleware
	authService      service.AuthService
	jwtSecretKey     []byte
)

func initAuthMiddleware(jwtSecret string) error {
	key, err := resolveJWTSecret(jwtSecret)
	if err != nil {
		return err
	}
	jwtSecretKey = key

	jm, err := jwt.New(initParams())
	if err != nil {
		return err
	}
	ginJwt = jm
	authService = service.GetAuthService()
	return nil
}

// resolveJWTSecret 获取 JWT 密钥，如果未配置则从数据库 DSN 派生稳定密钥。
func resolveJWTSecret(secret string) ([]byte, error) {
	if secret != "" {
		return []byte(secret), nil
	}
	dns := config.Get().DB.DNS
	if dns == "" {
		dns = "nexus"
	}
	hash := sha256.Sum256([]byte(dns + "_nexus_jwt_secret"))
	key := []byte(hex.EncodeToString(hash[:]))
	log.Warn().Msg("JWT 密钥未配置，已从数据库 DSN 派生稳定密钥。建议通过 server.jwt_secret 或 NEXUS_SERVER_JWT_SECRET 配置。")
	return key, nil
}

func Auth() gin.HandlerFunc {
	jwtMiddleware := ginJwt.MiddlewareFunc()
	return func(ctx *gin.Context) {
		token := extractToken(ctx)
		if token != "" && service.IsTokenBlacklisted(token) {
			response.Fail(ctx, errcode.ErrUnauthorized)
			ctx.Abort()
			return
		}
		jwtMiddleware(ctx)
	}
}

func initParams() *jwt.GinJWTMiddleware {
	return &jwt.GinJWTMiddleware{
		Key:             jwtSecretKey,
		IdentityKey:     constant.HttpHeaderUserIDKey,
		Timeout:         time.Hour * 24,     // JWT Token 的有效期
		MaxRefresh:      time.Hour * 24 * 7, // 刷新 Token 的有效期
		Authenticator:   authenticator(),    // 验证用户的回调函数。返回用户数据
		LoginResponse:   loginSuccess(),     // 登录成功后的回调函数
		Unauthorized:    unauthorized(),     // 处理未授权请求的回调函数
		IdentityHandler: identityHandler(),  // 从 JWT Token 中提取用户身份的回调函数
		PayloadFunc:     payloadFunc(),
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
	}
}

// extractToken 从请求中提取 Token 字符串
func extractToken(ctx *gin.Context) string {
	// 优先从 Authorization header 提取
	token := ctx.GetHeader("Authorization")
	if token != "" {
		return token
	}
	// 其次从 query 参数提取
	token = ctx.Query("token")
	if token != "" {
		return token
	}
	// 最后从 cookie 提取
	token, _ = ctx.Cookie("jwt")
	return token
}

// 登录
func authenticator() func(ctx *gin.Context) (any, error) {
	return func(ctx *gin.Context) (any, error) {
		var req request.LoginRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return nil, err
		}
		res, err := authService.Login(ctx, &req)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

// 登录成功
func loginSuccess() func(ctx *gin.Context, token *core.Token) {
	return func(ctx *gin.Context, token *core.Token) {
		response.Success(ctx, response.LoginResponse{
			Token:        token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.ExpiresAt,
		})
	}
}

// 登录失败
func unauthorized() func(ctx *gin.Context, code int, message string) {
	return func(ctx *gin.Context, code int, message string) {
		response.Fail(ctx, errcode.ErrUnauthorized)
	}
}

// 从 JWT Token 中提取用户身份
func identityHandler() func(ctx *gin.Context) any {
	return func(ctx *gin.Context) any {
		claims := jwt.ExtractClaims(ctx)
		return utils.StringToUint64(claims[constant.HttpHeaderUserIDKey].(string))
	}
}

// 登录处理器
func LoginHandler() gin.HandlerFunc {
	return ginJwt.LoginHandler
}

func payloadFunc() func(data any) gojwt.MapClaims {
	return func(data any) gojwt.MapClaims {
		if v, ok := data.(*model.User); ok {
			return gojwt.MapClaims{
				constant.HttpHeaderUserIDKey: utils.Uint64ToString(v.ID),
			}
		}
		return gojwt.MapClaims{}
	}
}
