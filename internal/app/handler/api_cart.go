package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// =====================
// REST API (лаб.3) — корзина (черновик)
// =====================

// ApiGetCartIcon — GET /api/brake-pad-wear/cart-icon
// Возвращает id черновой заявки и количество услуг в ней.
// @Summary Get cart icon info
// @Tags brake-pad-wear
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /brake-pad-wear/cart-icon [get]
func (h *Handler) ApiGetCartIcon(ctx *gin.Context) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
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

type addToCartRequest struct {
	ServiceID uint `json:"service_id" binding:"required"`
	Quantity  int  `json:"quantity"`
}

// ApiAddToCart — POST /api/brake-pad-wear/cart/items
// Добавляет услугу в заявку-черновик текущего пользователя.
// Если черновика нет — он создаётся автоматически.
// @Summary Add item to draft application
// @Tags brake-pad-wear
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param payload body addToCartRequest true "item payload"
// @Success 200 {object} repository.Application
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /brake-pad-wear/cart/items [post]
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

