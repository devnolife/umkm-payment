// Small one-shot tool to verify DB connectivity and report current table row counts.
// Run from apps/api: `go run ./cmd/dbcheck`
package main

import (
	"fmt"
	"log"

	"github.com/devnolife/umkm-api/internal/config"
	"github.com/devnolife/umkm-api/internal/database"
	"github.com/devnolife/umkm-api/internal/models"
)

func main() {
	config.Load()
	db := database.Connect()

	if err := db.AutoMigrate(
		&models.User{}, &models.Store{}, &models.Category{},
		&models.MenuItem{}, &models.Order{}, &models.OrderItem{}, &models.Payment{},
	); err != nil {
		log.Fatalf("automigrate: %v", err)
	}

	counts := map[string]int64{}
	tables := []struct {
		name string
		m    any
	}{
		{"users", &models.User{}},
		{"stores", &models.Store{}},
		{"categories", &models.Category{}},
		{"menu_items", &models.MenuItem{}},
		{"orders", &models.Order{}},
		{"payments", &models.Payment{}},
	}
	for _, t := range tables {
		var c int64
		if err := db.Model(t.m).Count(&c).Error; err != nil {
			log.Fatalf("count %s: %v", t.name, err)
		}
		counts[t.name] = c
	}

	fmt.Println("=== DB ROW COUNTS ===")
	for _, t := range tables {
		fmt.Printf("  %-12s %d\n", t.name, counts[t.name])
	}

	var admins, sellers, buyers int64
	db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&admins)
	db.Model(&models.User{}).Where("role = ?", models.RoleSeller).Count(&sellers)
	db.Model(&models.User{}).Where("role = ?", models.RoleBuyer).Count(&buyers)
	fmt.Println("=== USER BREAKDOWN ===")
	fmt.Printf("  admin:%d  seller:%d  buyer:%d\n", admins, sellers, buyers)
}
