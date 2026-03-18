package handler

import (
	"fmt"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// singletonUserID — фиксированный пользователь для всех REST-запросов (по ТЗ).
func singletonUserID() uint {
	return 1
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
		"app":      app,
		"minioBase": h.minioBase(ctx),
	})
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
		"UPDATE applications SET status = 'deleted' WHERE created_by_id = $1 AND status = 'draft'",
		userID,
	)
	if err != nil {
		logrus.WithError(err).Error("delete draft sql error")
	}

	ctx.Redirect(http.StatusFound, "/")
}

// =====================
// REST API (лаб.3) — все маршруты начинаются с /api/...
// =====================

// ApiGetServices — GET /api/services
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

// ApiGetService — GET /api/services/:id
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
	Title       string `json:"title"       binding:"required"`
	Description string `json:"description" binding:"required"`
	Status      string `json:"status"`
	ImageURL    string `json:"image_url"`
	VideoURL    string `json:"video_url"`
	PadType     string `json:"pad_type"    binding:"required"`
	BaseResource int   `json:"base_resource" binding:"required"`
	Price       int    `json:"price"       binding:"required"`
}

// ApiCreateService — POST /api/services
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

// ApiGetCartIcon — GET /api/cart-icon
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

// ApiAddToCart — POST /api/cart/items
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

// ApiGetApplications — GET /api/applications
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

// ApiGetApplication — GET /api/applications/:id
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

// ApiDeleteApplicationItem — DELETE /api/applications/:appId/items/:serviceId
func (h *Handler) ApiDeleteApplicationItem(ctx *gin.Context) {
	userID := singletonUserID()
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

// ApiUpdateApplicationItem — PUT /api/applications/:appId/items/:serviceId
func (h *Handler) ApiUpdateApplicationItem(ctx *gin.Context) {
	userID := singletonUserID()
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

type updateApplicationRequest struct {
	DrivingStyle string `json:"driving_style"`
	Mileage      *int   `json:"mileage"`
	Status       string `json:"status"`
}

// ApiUpdateApplication — PUT /api/applications/:id
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

// ApiApproveApplication — POST /api/applications/:id/approve
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

// ApiRejectApplication — POST /api/applications/:id/reject
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
// Аналог PUT /api/system_loads/{id}/finish из примера: переводит
// сформированную заявку в статус completed от имени модератора.
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
		// например, если статус не formed
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, app)
}

// ApiDeleteApplication — DELETE /api/applications/:id
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

// =====================
// Пользователь / аутентификация (заглушки для ЛР4)
// =====================

type registerUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	IsModerator bool   `json:"is_moderator"`
}

// ApiRegisterUser — POST /api/users/register
// Реальная авторизация пока не нужна, но регистрация создаёт запись в таблице users.
func (h *Handler) ApiRegisterUser(ctx *gin.Context) {
	var req registerUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// для простоты используем пароль как есть, без хеша (это ок для учебной заглушки)
	u, err := h.Repository.CreateUser(req.Username, req.Password, req.IsModerator)
	if err != nil {
		logrus.WithError(err).Error("api: register user error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot register user"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":          u.ID,
		"username":    u.Username,
		"isModerator": u.IsModerator,
	})
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ApiLogin — POST /api/auth/login
// Заглушка для 4-й лабораторной: не меняет singleton-пользователя, лишь возвращает успешный ответ.
func (h *Handler) ApiLogin(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// В реальном приложении здесь проверяли бы пароль и устанавливали cookie / токен.
	ctx.JSON(http.StatusOK, gin.H{
		"message":  "login stub OK",
		"username": req.Username,
	})
}

// ApiLogout — POST /api/auth/logout
// Заглушка: просто подтверждает выход.
func (h *Handler) ApiLogout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "logout stub OK",
	})
}