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

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("Ошибка инициализации репозитория")
	}

	h := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/", h.GetPads)
	// страница услуги: расчёт brake pad wear для выбранного типа колодок
	r.GET("/brake-pad/:id", h.GetPad)
	// страница заявки brake pad wear
	r.GET("/brake-pad-wear/:id", h.GetCalculation)

	// слушаем на всех интерфейсах, чтобы можно было открыть с телефона
	r.Run("0.0.0.0:8080")
	log.Println("Server down")
}
