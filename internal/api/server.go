package api

import (
	"log"

	"web_backend/internal/app/handler"
	"web_backend/internal/app/repository"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/sirupsen/logrus"

	_ "web_backend/docs"
)

// RegisterHandler godoc
// @title Brake Pad Wear API
// @version 1.0
// @description Лабораторная 4: авторизация (JWT+blacklist Redis) и Swagger для SPA.
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

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
	if err := verifyCoreSchema(sqlDB); err != nil {
		logrus.WithError(err).Fatal("DB schema mismatch — главная и API не смогут работать, пока не восстановите таблицы")
	}

	redisClient, err := connectRedis()
	if err != nil {
		logrus.WithError(err).Fatal("Redis connection error")
	}

	repo := repository.NewRepository(gormDB, redisClient)
	h := handler.NewHandler(repo, sqlDB)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// Swagger
	swaggerURL := ginSwagger.URL("/swagger/doc.json")
	r.Any("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(301, "/swagger/index.html")
	})

	// HTML-интерфейс (из предыдущих лабораторных)
	r.GET("/", h.GetServices)
	// новые человекочитаемые URL под тему brake pad
	r.GET("/brake-pad/:id", h.GetService)
	r.GET("/applications", h.GetApplicationsList)
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
		// Гость: регистрация/логин и чтение каталога
		api.POST("/users/register", h.ApiRegisterUser)
		api.POST("/auth/login", h.ApiLogin)

		api.GET("/brake-pad", h.ApiGetServices)
		api.GET("/brake-pad/:id", h.ApiGetService)

		// Авторизованный пользователь
		authorized := api.Group("/")
		authorized.Use(h.RequireAuth())
		authorized.POST("/auth/logout", h.ApiLogout)

		authorized.POST("/brake-pad", h.ApiCreateService)
		authorized.GET("/brake-pad-wear/cart-icon", h.ApiGetCartIcon)
		authorized.POST("/brake-pad-wear/cart/items", h.ApiAddToCart)

		authorized.GET("/brake-pad-wear", h.ApiGetApplications)
		authorized.GET("/brake-pad-wear/:id", h.ApiGetApplication)
		authorized.PUT("/brake-pad-wear/:id", h.ApiUpdateApplication)
		authorized.DELETE("/brake-pad-wear/:id", h.ApiDeleteApplication)

		authorized.PUT("/brake-pad-wear/:id/items/:serviceId", h.ApiUpdateApplicationItem)
		authorized.DELETE("/brake-pad-wear/:id/items/:serviceId", h.ApiDeleteApplicationItem)

		// Только модератор
		moderator := authorized.Group("/")
		moderator.Use(h.RequireModerator())
		moderator.PUT("/brake-pad-wear/:id/finish", h.ApiFinishApplication)
		moderator.PUT("/brake-pad-wear/:id/reject", h.ApiRejectApplication)
	}

	// слушаем на всех интерфейсах, чтобы можно было открыть с телефона
	r.Run(":8080")
	log.Println("Server down")
}
