package database

import (
	"log"

	"github.com/devnolife/umkm-api/internal/config"
	"github.com/devnolife/umkm-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() *gorm.DB {
	cfg := config.Get()

	logLevel := logger.Warn
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("[database] failed to connect: %v", err)
	}

	DB = db
	log.Println("[database] connected to PostgreSQL")
	return db
}

// AutoMigrate menjalankan GORM AutoMigrate untuk semua model.
// CATATAN: Skema sumber kebenaran tetap Prisma di apps/web. Gunakan ini untuk dev cepat saja.
func AutoMigrate() {
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Store{},
		&models.Category{},
		&models.MenuItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
	); err != nil {
		log.Fatalf("[database] migrate error: %v", err)
	}
	log.Println("[database] auto-migrate done")
}
