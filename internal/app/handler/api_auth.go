package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/golang-jwt/jwt/v5"
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
// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body registerUserRequest true "register payload"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /users/register [post]
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
// Возвращает JWT (Bearer) для остальных запросов.
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body loginRequest true "login payload"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /auth/login [post]
func (h *Handler) ApiLogin(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	u, err := h.Repository.GetUserByUsername(req.Username)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверный логин или пароль"})
		return
	}
	if u.PasswordHash != req.Password {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверный логин или пароль"})
		return
	}

	now := time.Now()
	claims := authClaims{
		UserID:      u.ID,
		IsModerator: u.IsModerator,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(h.jwtKey())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot sign token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": tokenStr,
		"user": gin.H{
			"id":          u.ID,
			"username":    u.Username,
			"isModerator": u.IsModerator,
		},
	})
}

// ApiLogout — POST /api/auth/logout
// Добавляет JWT в blacklist Redis до истечения срока.
// @Summary Logout (blacklist token)
// @Tags auth
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /auth/logout [post]
func (h *Handler) ApiLogout(ctx *gin.Context) {
	tokenStr := extractBearerToken(ctx)
	if tokenStr == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims := &authClaims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return h.jwtKey(), nil
	})
	if err != nil || tok == nil || !tok.Valid || claims.ExpiresAt == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err := h.Repository.BlacklistJWT(ctx, tokenStr, claims.ExpiresAt.Time); err != nil {
		logrus.WithError(err).Error("api: blacklist token error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "cannot logout"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

