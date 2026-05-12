package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const bearerPrefix = "Bearer "

type authClaims struct {
	UserID      uint `json:"user_id"`
	IsModerator bool `json:"is_moderator"`
	jwt.RegisteredClaims
}

func (h *Handler) jwtKey() []byte {
	key := os.Getenv("JWT_KEY")
	if key == "" {
		key = "default-secret-key-change-in-production"
	}
	return []byte(key)
}

func extractBearerToken(c *gin.Context) string {
	v := c.GetHeader("Authorization")
	if v == "" || !strings.HasPrefix(v, bearerPrefix) {
		return ""
	}
	return strings.TrimPrefix(v, bearerPrefix)
}

func (h *Handler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c)
		if tokenStr == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if h.Repository != nil && h.Repository.Redis != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			blacklisted, err := h.Repository.IsTokenBlacklisted(ctx, tokenStr)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			if blacklisted {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		claims := &authClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return h.jwtKey(), nil
		})
		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("is_moderator", claims.IsModerator)
		c.Next()
	}
}

func (h *Handler) RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("is_moderator")
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		isMod, _ := v.(bool)
		if !isMod {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

func userIDFromCtx(c *gin.Context) (uint, error) {
	v, ok := c.Get("user_id")
	if !ok || v == nil {
		return 0, fmt.Errorf("user_id not found")
	}
	switch t := v.(type) {
	case uint:
		return t, nil
	case int:
		if t <= 0 {
			return 0, fmt.Errorf("invalid user_id")
		}
		return uint(t), nil
	case float64:
		if t <= 0 {
			return 0, fmt.Errorf("invalid user_id")
		}
		return uint(t), nil
	case string:
		n, err := strconv.Atoi(t)
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid user_id")
		}
		return uint(n), nil
	default:
		return 0, fmt.Errorf("invalid user_id type")
	}
}

