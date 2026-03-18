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
func (h *Handler) ApiGetCartIcon(ctx *gin.Context) {
	userID := singletonUserID()
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
func (h *Handler) ApiAddToCart(ctx *gin.Context) {
	userID := singletonUserID()
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

