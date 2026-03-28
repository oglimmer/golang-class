package repository

import (
	"fmt"

	"github.com/oglimmer/kniffel/config"
	"github.com/oglimmer/kniffel/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates a new database connection and runs migrations
func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	logMode := logger.Warn
	if cfg.LogLevel == "debug" {
		logMode = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// NOTE: AutoMigrate is convenient for development, but for production
	// use a proper migration tool like golang-migrate/migrate or goose.
	// AutoMigrate cannot handle column renames, deletions, or complex schema changes safely.
	err = db.AutoMigrate(
		&model.GameEntity{},
		&model.PlayerEntity{},
		&model.DiceRollEntity{},
		&model.UsedBookingTypeEntity{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
