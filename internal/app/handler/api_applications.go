package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/repository"
)

// =====================
// REST API (лаб.3) — домен заявок (brake pad wear)
// =====================

// Плоские DTO без вложенного объекта service внутри строки (удобно для проверки и Postman).

type applicationLineFlat struct {
	ServiceID           uint    `json:"service_id"`
	ServiceTitle        string  `json:"service_title"`
	PadType             string  `json:"pad_type"`
	BaseResourceKm      int     `json:"base_resource_km"`
	DrivingStyleCoeff   float64 `json:"driving_style_coefficient"`
	EffectiveResourceKm float64 `json:"effective_resource_km"`
	MileageKm           int     `json:"mileage_km"`
	Quantity            int     `json:"quantity"`
	RemainingKm         float64 `json:"remaining_km"`
	RemainingPercent    int     `json:"remaining_percent"`
	Comment             string  `json:"comment"`
}

type applicationFlat struct {
	ID                  uint                  `json:"id"`
	Status              string                `json:"status"`
	CreatedAt           time.Time             `json:"created_at"`
	FormedAt            *time.Time            `json:"formed_at,omitempty"`
	CompletedAt         *time.Time            `json:"completed_at,omitempty"`
	DrivingStyle        string                `json:"driving_style"`
	Mileage             int                   `json:"mileage"`
	DrivingStyleCoeff   float64               `json:"driving_style_coefficient"`
	TotalPrice          int                   `json:"total_price"`
	ItemsCount          int                   `json:"items_count"`
	MinRemainingKm      float64               `json:"min_remaining_km"`
	MinRemainingPercent int                   `json:"min_remaining_percent"`
	Lines               []applicationLineFlat `json:"lines"`
}

// applicationSummaryJSON — укороченный формат списка заявок, без вложенных Items/Service,
// аналогичный примеру system_loads из методички.
type applicationSummaryJSON struct {
	ID                uint       `json:"id"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	FormedAt          *time.Time `json:"formed_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	DrivingStyle      string     `json:"driving_style"`
	Mileage           int        `json:"mileage"`
	DrivingStyleCoeff float64    `json:"driving_style_coefficient"`
	MinRemainingKm    float64    `json:"min_remaining_km"`
}

type applicationsListResponse struct {
	Total   int                      `json:"total"`
	Results []applicationSummaryJSON `json:"results"`
}

func completedMinRemainingKm(app repository.Application) float64 {
	if app.Status != "completed" {
		return 0
	}
	return app.MinRemainingKm
}

func buildApplicationFlat(app *repository.Application) applicationFlat {
	lines := make([]applicationLineFlat, 0, len(app.Items))
	for _, it := range app.Items {
		lines = append(lines, applicationLineFlat{
			ServiceID:           it.ServiceID,
			ServiceTitle:        it.Service.Title,
			PadType:             it.Service.PadType,
			BaseResourceKm:      it.Service.BaseResource,
			DrivingStyleCoeff:   app.DrivingStyleCoeff,
			EffectiveResourceKm: it.EffectiveResourceKm,
			MileageKm:           app.Mileage,
			Quantity:            it.Quantity,
			RemainingKm:         it.RemainingKM,
			RemainingPercent:    it.RemainingPercent,
			Comment:             it.Comment,
		})
	}
	return applicationFlat{
		ID:                  app.ID,
		Status:              app.Status,
		CreatedAt:           app.CreatedAt,
		FormedAt:            app.FormedAt,
		CompletedAt:         app.CompletedAt,
		DrivingStyle:        app.DrivingStyle,
		Mileage:             app.Mileage,
		DrivingStyleCoeff:   app.DrivingStyleCoeff,
		TotalPrice:          app.TotalPrice,
		ItemsCount:          len(app.Items),
		MinRemainingKm:      app.MinRemainingKm,
		MinRemainingPercent: app.MinRemainingPercent,
		Lines:               lines,
	}
}

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
	out := make([]applicationSummaryJSON, 0, len(apps))
	for i := range apps {
		a := apps[i]
		out = append(out, applicationSummaryJSON{
			ID:                a.ID,
			Status:            a.Status,
			CreatedAt:         a.CreatedAt,
			FormedAt:          a.FormedAt,
			CompletedAt:       a.CompletedAt,
			DrivingStyle:      a.DrivingStyle,
			Mileage:           a.Mileage,
			DrivingStyleCoeff: a.DrivingStyleCoeff,
			MinRemainingKm:    completedMinRemainingKm(a),
		})
	}
	ctx.JSON(http.StatusOK, applicationsListResponse{
		Total:   len(out),
		Results: out,
	})
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
	ctx.JSON(http.StatusOK, buildApplicationFlat(app))
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

	_, err = h.Repository.UpdateApplicationForCreator(uint(id), userID, fields)
	if err != nil {
		logrus.WithError(err).Error("api: update application error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update application"})
		return
	}
	app, err := h.Repository.GetApplicationByID(uint(id), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot reload application"})
		return
	}
	ctx.JSON(http.StatusOK, buildApplicationFlat(app))
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
