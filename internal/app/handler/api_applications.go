package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// =====================
// REST API (лаб.3) — домен заявок (brake pad wear)
// =====================

// ApiGetApplications — GET /api/brake-pad-wear
// Список заявок (кроме удалённых и черновиков) с фильтрацией по статусу
// и диапазону даты формирования.
func (h *Handler) ApiGetApplications(ctx *gin.Context) {
	userID := singletonUserID()
	status := ctx.Query("status")
	fromStr := ctx.Query("formed_from")
	toStr := ctx.Query("formed_to")

	var fromPtr, toPtr *time.Time
	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			fromPtr = &t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			toPtr = &t
		}
	}

	apps, err := h.Repository.FilterApplicationsForUser(userID, status, fromPtr, toPtr)
	if err != nil {
		logrus.WithError(err).Error("api: applications list error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot load applications"})
		return
	}
	ctx.JSON(http.StatusOK, apps)
}

// ApiGetApplication — GET /api/brake-pad-wear/:id
func (h *Handler) ApiGetApplication(ctx *gin.Context) {
	userID := singletonUserID()
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	app, err := h.Repository.GetApplicationByID(uint(id), userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return
	}
	ctx.JSON(http.StatusOK, app)
}

type updateApplicationRequest struct {
	DrivingStyle string `json:"driving_style"`
	Mileage      *int   `json:"mileage"`
	Status       string `json:"status"`
}

// ApiUpdateApplication — PUT /api/brake-pad-wear/:id
func (h *Handler) ApiUpdateApplication(ctx *gin.Context) {
	userID := singletonUserID()
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateApplicationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	fields := map[string]any{}
	if req.DrivingStyle != "" {
		fields["driving_style"] = req.DrivingStyle
	}
	if req.Mileage != nil {
		fields["mileage"] = *req.Mileage
	}
	if req.Status != "" {
		fields["status"] = req.Status
	}
	if len(fields) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	app, err := h.Repository.UpdateApplicationForCreator(uint(id), userID, fields)
	if err != nil {
		logrus.WithError(err).Error("api: update application error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update application"})
		return
	}
	ctx.JSON(http.StatusOK, app)
}

// ApiApproveApplication — POST /api/brake-pad-wear/:id/approve
func (h *Handler) ApiApproveApplication(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// в учебной работе считаем, что модератор с id=2
	const moderatorID = uint(2)
	app, err := h.Repository.SetApplicationStatus(uint(id), "completed", moderatorID)
	if err != nil {
		logrus.WithError(err).Error("api: approve application error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, app)
}

// ApiRejectApplication — POST /api/brake-pad-wear/:id/reject
func (h *Handler) ApiRejectApplication(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	const moderatorID = uint(2)
	app, err := h.Repository.SetApplicationStatus(uint(id), "rejected", moderatorID)
	if err != nil {
		logrus.WithError(err).Error("api: reject application error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, app)
}

// ApiFinishApplication — PUT /api/brake-pad-wear/:id/finish
// Аналог PUT /api/.../{id}/finish из примера: переводит сформированную заявку
// в статус completed от имени модератора.
func (h *Handler) ApiFinishApplication(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	const moderatorID = uint(2)
	app, err := h.Repository.SetApplicationStatus(uint(id), "completed", moderatorID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, app)
}

// ApiDeleteApplication — DELETE /api/brake-pad-wear/:id
func (h *Handler) ApiDeleteApplication(ctx *gin.Context) {
	userID := singletonUserID()
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.Repository.SoftDeleteApplication(uint(id), userID); err != nil {
		logrus.WithError(err).Error("api: delete application error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete application"})
		return
	}
	ctx.Status(http.StatusNoContent)
}

