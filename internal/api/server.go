package api

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"web_backend/internal/app/handler"
	"web_backend/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	r.GET("/", h.GetServices)
	// новые URL только в стиле brake pad
	r.GET("/brake-pad/:id", h.GetService)
	r.GET("/brake-pad-wear/:id", h.GetApplication)
	r.POST("/brake-pad-wear/:id/recalc", h.RecalcWear)

	r.POST("/cart/add", h.AddToDraft)
	r.POST("/cart/delete", h.DeleteDraft)

	// слушаем на всех интерфейсах, чтобы можно было открыть с телефона
	r.Run(":8080")
	log.Println("Server down")
}

func connectDB() (*sql.DB, *gorm.DB, error) {
	host := "127.0.0.1"
	port := getenv("DB_PORT", "5433")
	user := getenv("DB_USER", "postgres")
	password := "postgres"
	dbname := getenv("DB_NAME", "mydb")
	sslmode := getenv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
	fmt.Println(dsn)

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return sqlDB, gormDB, nil
}

func runMigrations(db *sql.DB) error {
	goose.SetDialect("postgres")
	return goose.Up(db, "migrations")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
