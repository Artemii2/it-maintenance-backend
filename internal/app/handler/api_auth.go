package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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

