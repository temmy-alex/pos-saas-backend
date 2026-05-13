package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"pos-saas-backend/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresConnection(databaseConfig config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		databaseConfig.Host,
		databaseConfig.Port,
		databaseConfig.Username,
		databaseConfig.Password,
		databaseConfig.Database,
		databaseConfig.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Println("Database connected successfully")

	return db, nil
}
