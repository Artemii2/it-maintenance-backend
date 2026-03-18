package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/repository"
)

// =====================
// REST API (лаб.3) — домен услуг (brake pad)
// =====================

// ApiGetServices — GET /api/brake-pad
// Список услуг с фильтрами:
//   - title   — подстрока в названии/описании
//   - padType — точное совпадение типа колодок
//   - status  — статус (по умолчанию все кроме deleted)
func (h *Handler) ApiGetServices(ctx *gin.Context) {
	title := ctx.Query("title")
	padType := ctx.Query("padType")
	status := ctx.Query("status")

	services, err := h.Repository.FilterServices(title, padType, status)
	if err != nil {
		logrus.WithError(err).Error("api: services filter error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot load services"})
		return
	}

	ctx.JSON(http.StatusOK, services)
}

// ApiGetService — GET /api/brake-pad/:id
func (h *Handler) ApiGetService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	svc, err := h.Repository.GetService(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	ctx.JSON(http.StatusOK, svc)
}

type createServiceRequest struct {
	Title        string `json:"title"       binding:"required"`
	Description  string `json:"description" binding:"required"`
	Status       string `json:"status"`
	ImageURL     string `json:"image_url"`
	VideoURL     string `json:"video_url"`
	PadType      string `json:"pad_type"      binding:"required"`
	BaseResource int    `json:"base_resource" binding:"required"`
	Price        int    `json:"price"         binding:"required"`
}

// ApiCreateService — POST /api/brake-pad
// Ожидает JSON с данными услуги. Поля image_url и video_url содержат ИМЕНА файлов,
// которые заранее загружены в MinIO.
func (h *Handler) ApiCreateService(ctx *gin.Context) {
	var req createServiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	svc := &repository.Service{
		Title:        req.Title,
		Description:  req.Description,
		Status:       req.Status,
		ImageURL:     req.ImageURL,
		VideoURL:     req.VideoURL,
		PadType:      req.PadType,
		BaseResource: req.BaseResource,
		Price:        req.Price,
	}

	if err := h.Repository.CreateService(svc); err != nil {
		logrus.WithError(err).Error("api: create service error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create service"})
		return
	}

	ctx.JSON(http.StatusCreated, svc)
}

