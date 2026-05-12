package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// =====================
// REST API (лаб.3) — корзина (черновик)
// =====================

// ApiGetCartIcon — GET /api/brake-wear/cart-icon
// Возвращает id черновой заявки и количество услуг в ней.
// @Summary Get cart icon info
// @Tags brake-wear
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /brake-wear/cart-icon [get]
func (h *Handler) ApiGetCartIcon(ctx *gin.Context) {
	userID := h.cartIconUserID(ctx)
	app, count, err := h.Repository.GetCartInfo(userID)
	if err != nil {
		logrus.WithError(err).Error("api: cart icon error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot load cart"})
		return
	}
	if app == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"application_id": nil,
			"items_count":    0,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"application_id": app.ID,
		"items_count":    count,
	})
}

func (h *Handler) cartIconUserID(ctx *gin.Context) uint {
	if userID, err := userIDFromCtx(ctx); err == nil {
		return userID
	}

	tokenStr := extractBearerToken(ctx)
	if tokenStr == "" {
		return 1
	}

	if h.Repository != nil && h.Repository.Redis != nil {
		redisCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		blacklisted, err := h.Repository.IsTokenBlacklisted(redisCtx, tokenStr)
		if err != nil || blacklisted {
			return 1
		}
	}

	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return h.jwtKey(), nil
	})
	if err != nil || token == nil || !token.Valid || claims.UserID == 0 {
		return 1
	}
	return claims.UserID
}

type addToCartRequest struct {
	ServiceID uint `json:"service_id" binding:"required"`
	Quantity  int  `json:"quantity"`
}

// ApiAddToCart — POST /api/brake-wear/cart/items
// Добавляет услугу в заявку-черновик текущего пользователя.
// Если черновика нет — он создаётся автоматически.
// @Summary Add item to draft application
// @Tags brake-wear
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param payload body addToCartRequest true "item payload"
// @Success 200 {object} repository.Application
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /brake-wear/cart/items [post]
func (h *Handler) ApiAddToCart(ctx *gin.Context) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var req addToCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || req.ServiceID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}

	app, err := h.Repository.AddServiceToDraft(userID, req.ServiceID, qty)
	if err != nil {
		logrus.WithError(err).Error("api: add to cart error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot add to cart"})
		return
	}
	ctx.JSON(http.StatusOK, app)
}
