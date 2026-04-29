package middleware

import (
	"time"

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
)

var (
	ginJwt      *jwt.GinJWTMiddleware
	authService service.AuthService
)

func initAuthMiddleware() error {
	jm, err := jwt.New(initParams())
	if err != nil {
		return err
	}
	ginJwt = jm
	authService = service.GetAuthService()
	return nil
}

func Auth() gin.HandlerFunc {
	return ginJwt.MiddlewareFunc()
}

func initParams() *jwt.GinJWTMiddleware {
	return &jwt.GinJWTMiddleware{
		Key:             []byte("secret key"),
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
		response.Fail(ctx, errcode.ErrLoginFailed)
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
