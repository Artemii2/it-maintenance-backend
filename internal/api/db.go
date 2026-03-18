package api

import (
	"database/sql"
	"fmt"
	"os"

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

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

