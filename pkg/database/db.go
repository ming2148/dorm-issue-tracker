package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dorm-issue-tracker/internal/domain"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() (*gorm.DB, error) {
	driver := os.Getenv("DB_DRIVER")
	var dialector gorm.Dialector

	if driver == "postgres" {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Bangkok",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", "postgres"),
			getEnv("DB_NAME", "dorm_tracker"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_SSLMODE", "disable"),
		)
		dialector = postgres.Open(dsn)
		log.Println("Connecting to PostgreSQL database...")
	} else {
		// Pure Go SQLite for immediate zero-config execution
		dbDir := getEnv("DB_DIR", "./data")
		_ = os.MkdirAll(dbDir, 0755)
		dbPath := filepath.Join(dbDir, "dorm.db")
		dialector = sqlite.Open(dbPath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
		log.Printf("Connecting to SQLite database at %s ...\n", dbPath)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto-Migrate Schemas
	log.Println("Running AutoMigrations...")
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Ticket{},
		&domain.TicketImage{},
		&domain.Feedback{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed successfully.")
	return db, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
