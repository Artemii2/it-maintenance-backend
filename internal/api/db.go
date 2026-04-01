package api

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib"
)

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

// verifyCoreSchema проверяет, что после goose в БД действительно есть таблицы приложения.
// Если версия в goose_db_version «обогнала» реальную схему (например, заменили файлы миграций),
// сервер раньше поднимался, а главная страница оставалась пустой из‑за ERROR relation ... does not exist.
func verifyCoreSchema(db *sql.DB) error {
	var ok bool
	// to_regclass учитывает кавычки в имени "brake-pad"
	if err := db.QueryRow(`SELECT to_regclass('public."brake-pad"') IS NOT NULL`).Scan(&ok); err != nil {
		return fmt.Errorf("schema check: %w", err)
	}
	if !ok {
		return fmt.Errorf(`таблица "brake-pad" отсутствует, хотя миграции goose помечены выполненными — схема БД не совпадает с проектом. Сбросьте данные Postgres и примените миграции заново, например: docker compose down -v && docker compose up -d postgres, затем перезапустите приложение`)
	}
	if err := db.QueryRow(`SELECT to_regclass('public."brake-pad-wear"') IS NOT NULL`).Scan(&ok); err != nil {
		return fmt.Errorf("schema check: %w", err)
	}
	if !ok {
		return fmt.Errorf(`таблица "brake-pad-wear" отсутствует — выполните полный прогон миграций на чистой БД (см. сообщение выше про docker compose down -v)`)
	}
	return nil
}

func connectRedis() (*redis.Client, error) {
	addr := getenv("REDIS_ADDR", "127.0.0.1:6379")
	password := getenv("REDIS_PASSWORD", "")
	dbStr := getenv("REDIS_DB", "0")
	dbNum := 0
	if n, err := strconv.Atoi(dbStr); err == nil && n >= 0 {
		dbNum = n
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNum,
	})

	// проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

