package middleware

import (
	"net/http"
	"strings"

	appErrors "github.com/HieuNguyenVV/sp-agent/internal/errors"
	"github.com/HieuNguyenVV/sp-agent/internal/log"
	"github.com/HieuNguyenVV/sp-agent/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	RefreshTokenHeader  = "RefreshToken"
	BearerPrefix        = "Bearer"
	CtxClaimsKey        = "claims"
	CtxUserIDKey        = "user_id"
	CtxRefreshTokenKey  = "refresh_token"
)

func NewAuthMiddleware(_ log.Logger, jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(AuthorizationHeader)
		if token == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		token = strings.TrimPrefix(token, BearerPrefix+" ")
		if token == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			handleAuthError(c, err)
			return
		}
		if claims == nil || claims.Subject == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		c.Set(CtxClaimsKey, claims)
		c.Set(CtxUserIDKey, claims.Subject)
		c.Next()
	}
}

func NewRefreshTokenMiddleware(_ log.Logger, jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(RefreshTokenHeader)
		if token == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		userID, err := jwtService.ValidateRefreshToken(token)
		if err != nil {
			handleAuthError(c, err)
			return
		}
		if userID == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		c.Set(CtxUserIDKey, userID)
		c.Set(CtxRefreshTokenKey, token)
		c.Next()
	}
}

func handleAuthError(ctx *gin.Context, err error) {
	handleError(ctx, err, appErrors.ErrUnauthorized)
}

func handleError(ctx *gin.Context, err error, defaultAppError appErrors.AppError) {
	if err == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, defaultAppError)
		return
	}
	_ = ctx.Error(err)
	if appError, ok := err.(appErrors.AppError); ok {
		ctx.AbortWithStatusJSON(appError.StatusCode, err)
		return
	}
	wrapped := defaultAppError.Wrap(err)
	ctx.AbortWithStatusJSON(wrapped.StatusCode, wrapped)
}
