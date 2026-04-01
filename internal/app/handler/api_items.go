package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// =====================
// REST API (лаб.3) — домен m-m заявки-услуги (таблица brake_wear)
// =====================

// ApiDeleteApplicationItem — DELETE /api/brake-pad-wear/:id/items/:serviceId
func (h *Handler) ApiDeleteApplicationItem(ctx *gin.Context) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	appID, err1 := strconv.Atoi(ctx.Param("id"))
	svcID, err2 := strconv.Atoi(ctx.Param("serviceId"))
	if err1 != nil || err2 != nil || appID <= 0 || svcID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid ids"})
		return
	}
	if err := h.Repository.DeleteApplicationItem(uint(appID), uint(svcID), userID); err != nil {
		logrus.WithError(err).Error("api: delete application item error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete item"})
		return
	}
	ctx.Status(http.StatusNoContent)
}

type updateItemRequest struct {
	Quantity  *int  `json:"quantity"`
	Position  *int  `json:"position"`
	IsPrimary *bool `json:"is_primary"`
}

// ApiUpdateApplicationItem — PUT /api/brake-pad-wear/:id/items/:serviceId
func (h *Handler) ApiUpdateApplicationItem(ctx *gin.Context) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	appID, err1 := strconv.Atoi(ctx.Param("id"))
	svcID, err2 := strconv.Atoi(ctx.Param("serviceId"))
	if err1 != nil || err2 != nil || appID <= 0 || svcID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid ids"})
		return
	}

	var req updateItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.Repository.UpdateApplicationItem(
		uint(appID), uint(svcID), userID, req.Quantity, req.Position, req.IsPrimary,
	); err != nil {
		logrus.WithError(err).Error("api: update application item error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update item"})
		return
	}
	ctx.Status(http.StatusNoContent)
}

