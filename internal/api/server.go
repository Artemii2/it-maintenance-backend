package api

import (
	"log"

	"web_backend/internal/app/handler"
	"web_backend/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// StartServer инициализирует репозиторий, обработчики и запускает HTTP-сервер.
func StartServer() {
	log.Println("Starting server")

	sqlDB, gormDB, err := connectDB()
	if err != nil {
		logrus.WithError(err).Fatal("DB connection error")
	}

	if err := runMigrations(sqlDB); err != nil {
		logrus.WithError(err).Fatal("DB migration error")
	}

	repo := repository.NewRepository(gormDB)
	h := handler.NewHandler(repo, sqlDB)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// HTML-интерфейс (из предыдущих лабораторных)
	r.GET("/", h.GetServices)
	// новые человекочитаемые URL под тему brake pad
	r.GET("/brake-pad/:id", h.GetService)
	r.GET("/brake-pad-wear/:id", h.GetApplication)
	r.POST("/brake-pad-wear/:id/recalc", h.RecalcWear)
	// старые URL тоже оставляем, чтобы не ломать старые ссылки
	r.GET("/services/:id", h.GetService)
	r.GET("/applications/:id", h.GetApplication)
	r.POST("/cart/add", h.AddToDraft)
	r.POST("/cart/delete", h.DeleteDraft)

	// REST API для SPA (лаб.3)
	api := r.Group("/api")
	{
		// домен услуг (brake pad)
		api.GET("/brake-pad", h.ApiGetServices)
		api.GET("/brake-pad/:id", h.ApiGetService)
		api.POST("/brake-pad", h.ApiCreateService)

		// домен заявок brake pad wear и корзины
		api.GET("/brake-pad-wear/cart-icon", h.ApiGetCartIcon)
		api.POST("/brake-pad-wear/cart/items", h.ApiAddToCart)

		api.GET("/brake-pad-wear", h.ApiGetApplications)
		api.GET("/brake-pad-wear/:id", h.ApiGetApplication)
		api.PUT("/brake-pad-wear/:id", h.ApiUpdateApplication)
		// завершение сформированной заявки (аналог finish)
		api.PUT("/brake-pad-wear/:id/finish", h.ApiFinishApplication)
		// отклонение сформированной заявки модератором
		api.PUT("/brake-pad-wear/:id/reject", h.ApiRejectApplication)
		api.DELETE("/brake-pad-wear/:id", h.ApiDeleteApplication)

		// домен м-м заявки-услуги brake pad wear
		api.PUT("/brake-pad-wear/:id/items/:serviceId", h.ApiUpdateApplicationItem)
		api.DELETE("/brake-pad-wear/:id/items/:serviceId", h.ApiDeleteApplicationItem)

		// домен пользователя / аутентификации (заглушки для ЛР4)
		api.POST("/users/register", h.ApiRegisterUser)
		api.POST("/auth/login", h.ApiLogin)
		api.POST("/auth/logout", h.ApiLogout)
	}

	// слушаем на всех интерфейсах, чтобы можно было открыть с телефона
	r.Run(":8080")
	log.Println("Server down")
}
