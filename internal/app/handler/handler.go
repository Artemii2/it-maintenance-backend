package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/repository"
)

// Handler содержит зависимости HTTP-обработчиков.
type Handler struct {
	Repository *repository.Repository
}

// NewHandler создаёт новый Handler с переданным репозиторием.
func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// GetPads — главная страница: список типов колодок + карточка заявки
func (h *Handler) GetPads(ctx *gin.Context) {
	var pads []repository.BrakePad
	var err error

	searchQuery := ctx.Query("query")

	if searchQuery == "" {
		pads, err = h.Repository.GetPads()
	} else {
		pads, err = h.Repository.GetPadsByTitle(searchQuery)
	}

	if err != nil {
		logrus.Error(err)
	}

	calculations, err := h.Repository.GetCalculations()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"strategies":   pads, // оставляем имя чтобы не ломать шаблон
		"query":        searchQuery,
		"calculations": calculations,
	})
}

// GetPad — страница конкретного типа колодок
func (h *Handler) GetPad(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		return
	}

	pad, err := h.Repository.GetPad(id)
	if err != nil {
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "strategy.html", gin.H{
		"strategy": pad,
	})
}

// GetCalculation — страница расчёта остаточного ресурса
func (h *Handler) GetCalculation(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		return
	}

	calc, err := h.Repository.GetCalculation(id)
	if err != nil {
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "calculation.html", gin.H{
		"calc": calc,
	})
}