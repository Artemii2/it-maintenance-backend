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

