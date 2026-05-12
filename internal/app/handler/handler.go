package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/repository"
)

// Handler содержит зависимости HTTP-обработчиков.
type Handler struct {
	Repository *repository.Repository
	SQL        *sql.DB
}

// NewHandler создаёт новый Handler с переданным репозиторием.
func NewHandler(r *repository.Repository, sqlDB *sql.DB) *Handler {
	return &Handler{
		Repository: r,
		SQL:        sqlDB,
	}
}

func (h *Handler) currentUserID(ctx *gin.Context) uint {
	if v, err := ctx.Cookie("uid"); err == nil && v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return uint(n)
		}
	}

	// демо-авторизация: фиксируем пользователя 1
	ctx.SetCookie("uid", "1", 60*60*24*365, "/", "", false, true)
	return 1
}

func (h *Handler) minioBase(ctx *gin.Context) string {
	host := ctx.Request.Host // например 192.168.1.10:8080
	hostOnly := strings.Split(host, ":")[0]
	return fmt.Sprintf("http://%s:9000", hostOnly)
}

// =====================
// HTML-обработчики (лаб.1-2)
// =====================

// GetServices — главная страница: список услуг + карточка текущей заявки (черновик)
func (h *Handler) GetServices(ctx *gin.Context) {
	userID := h.currentUserID(ctx)

	searchQuery := ctx.Query("query")

	services, err := h.Repository.SearchServices(searchQuery)
	if err != nil {
		logrus.WithError(err).Error("services search error")
	}

	draft, err := h.Repository.GetDraftApplication(userID)
	if err != nil {
		logrus.WithError(err).Error("draft load error")
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"services":  services,
		"query":     searchQuery,
		"draft":     draft,
		"minioBase": h.minioBase(ctx),
	})
}

// GetApplicationsList — HTML: все заявки пользователя (кроме черновика и удалённых), плоская таблица.
func (h *Handler) GetApplicationsList(ctx *gin.Context) {
	userID := h.currentUserID(ctx)
	apps, err := h.Repository.FilterApplicationsForUser(userID, "", nil, nil)
	if err != nil {
		logrus.WithError(err).Error("applications list page error")
		apps = nil
	}
	ctx.HTML(http.StatusOK, "brake-pad-list.html", gin.H{
		"total":     len(apps),
		"apps":      apps,
		"minioBase": h.minioBase(ctx),
	})
}

// GetService — страница конкретной услуги
func (h *Handler) GetService(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		return
	}

	service, err := h.Repository.GetService(uint(id))
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.HTML(http.StatusOK, "brake-pad.html", gin.H{
		"strategy":  service,
		"minioBase": h.minioBase(ctx),
	})
}

// GetApplication — страница заявки (черновик/сформированная/...)
func (h *Handler) GetApplication(ctx *gin.Context) {
	userID := h.currentUserID(ctx)
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		return
	}

	app, err := h.Repository.GetApplicationByID(uint(id), userID)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.HTML(http.StatusOK, "brake-pad-ware.html", gin.H{
		"app":       app,
		"minioBase": h.minioBase(ctx),
	})
}

// RecalcWear — пересчитать износ (обновить стиль/пробег) и вернуть на страницу заявки.
func (h *Handler) RecalcWear(ctx *gin.Context) {
	userID := h.currentUserID(ctx)
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.Status(http.StatusBadRequest)
		return
	}

	style := strings.TrimSpace(ctx.PostForm("driving_style"))
	mileageStr := strings.TrimSpace(ctx.PostForm("mileage"))
	mileage := 0
	if mileageStr != "" {
		if n, err := strconv.Atoi(mileageStr); err == nil && n >= 0 {
			mileage = n
		}
	}

	if style == "" {
		style = "Спортивный"
	}

	if err := h.Repository.UpdateApplicationWearParams(uint(id), userID, style, mileage); err != nil {
		logrus.WithError(err).Error("recalc wear error")
	}

	ctx.Redirect(http.StatusFound, fmt.Sprintf("/brake-wear/%d", id))
}

// AddToDraft — добавление услуги в текущую заявку через ORM
func (h *Handler) AddToDraft(ctx *gin.Context) {
	userID := h.currentUserID(ctx)

	serviceIDStr := ctx.PostForm("service_id")
	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil || serviceID <= 0 {
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	_, err = h.Repository.AddServiceToDraft(userID, uint(serviceID), 1)
	if err != nil {
		logrus.WithError(err).Error("add to draft error")
	}

	ctx.Redirect(http.StatusFound, "/")
}

// DeleteDraft — логическое удаление заявки через SQL UPDATE (без ORM)
func (h *Handler) DeleteDraft(ctx *gin.Context) {
	userID := h.currentUserID(ctx)

	_, err := h.SQL.Exec(
		`UPDATE "brake-wear" SET status = 'deleted' WHERE created_by_id = $1 AND status = 'draft'`,
		userID,
	)
	if err != nil {
		logrus.WithError(err).Error("delete draft sql error")
	}

	ctx.Redirect(http.StatusFound, "/")
}
