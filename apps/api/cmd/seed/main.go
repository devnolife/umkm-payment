// Idempotent seed for development.
//
// Run:
//   cd apps/api && go run ./cmd/seed
//   or via: pnpm api:seed
//
// Seeds default accounts:
//
//   role     | username    | password   | notes
//   ---------+-------------+------------+---------------------------------
//   ADMIN    | admin       | admin123   | web admin dashboard
//   SELLER   | warungbudi  | seller123  | owner of "Warung Budi"
//   SELLER   | kantinmaya  | seller123  | owner of "Kantin Maya"
//   BUYER    | pembeli1    | buyer123   | for web buyer flow
//   BUYER    | pembeli2    | buyer123   | for web buyer flow
//   BUYER    | mobileuser  | mobile123  | for Expo mobile app
//
// Re-running is safe — users / stores / categories / menu items are upserted
// by their unique keys (username, sellerId, store+category name, etc.).
package main

import (
	"fmt"
	"log"

	"github.com/devnolife/umkm-api/internal/config"
	"github.com/devnolife/umkm-api/internal/database"
	"github.com/devnolife/umkm-api/internal/models"
	"github.com/devnolife/umkm-api/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type seedUser struct {
	Username string
	Name     string
	Password string
	Role     models.UserRole
	Email    string
	Phone    string
}

type seedMenu struct {
	Category string
	Name     string
	Desc     string
	Price    int
}

type seedStore struct {
	SellerUsername string
	Name           string
	Description    string
	Address        string
	Phone          string
	OpenTime       string
	CloseTime      string
	Categories     []string
	Menu           []seedMenu
}

func main() {
	config.Load()
	db := database.Connect()

	if err := db.AutoMigrate(
		&models.User{}, &models.Store{}, &models.Category{},
		&models.MenuItem{}, &models.Order{}, &models.OrderItem{}, &models.Payment{},
	); err != nil {
		log.Fatalf("[seed] automigrate: %v", err)
	}

	users := []seedUser{
		{"admin", "Administrator", "admin123", models.RoleAdmin, "admin@umkm.local", "081200000000"},
		{"warungbudi", "Budi Santoso", "seller123", models.RoleSeller, "budi@umkm.local", "081200000101"},
		{"kantinmaya", "Maya Putri", "seller123", models.RoleSeller, "maya@umkm.local", "081200000102"},
		{"pembeli1", "Pembeli Satu", "buyer123", models.RoleBuyer, "pembeli1@umkm.local", "081200000201"},
		{"pembeli2", "Pembeli Dua", "buyer123", models.RoleBuyer, "pembeli2@umkm.local", "081200000202"},
		{"mobileuser", "Andi Mobile", "mobile123", models.RoleBuyer, "andi@umkm.local", "081200000203"},
	}
	userIDByUsername := map[string]string{}
	for _, u := range users {
		id, err := upsertUser(db, u)
		if err != nil {
			log.Fatalf("[seed] user %s: %v", u.Username, err)
		}
		userIDByUsername[u.Username] = id
		fmt.Printf("  user  %-12s %-7s id=%s\n", u.Username, u.Role, id)
	}

	stores := []seedStore{
		{
			SellerUsername: "warungbudi",
			Name:           "Warung Budi",
			Description:    "Masakan rumahan khas Jawa Tengah, porsi mengenyangkan.",
			Address:        "Jl. Melati No. 12, Semarang",
			Phone:          "081200000101",
			OpenTime:       "08:00",
			CloseTime:      "21:00",
			Categories:     []string{"Nasi", "Lauk", "Minuman"},
			Menu: []seedMenu{
				{"Nasi", "Nasi Goreng Spesial", "Nasi goreng dengan telur, ayam, dan kerupuk.", 18000},
				{"Nasi", "Nasi Rames", "Nasi dengan tiga macam lauk pilihan.", 15000},
				{"Lauk", "Ayam Goreng Lengkuas", "Ayam goreng kremes bumbu lengkuas.", 12000},
				{"Lauk", "Tempe Mendoan", "Tempe iris tipis goreng basah.", 6000},
				{"Minuman", "Es Teh Manis", "Teh tubruk gula merah.", 4000},
				{"Minuman", "Es Jeruk", "Jeruk peras segar.", 7000},
			},
		},
		{
			SellerUsername: "kantinmaya",
			Name:           "Kantin Maya",
			Description:    "Snack, mie, dan kopi nikmat untuk anak kuliahan.",
			Address:        "Jl. Anggrek No. 5, Yogyakarta",
			Phone:          "081200000102",
			OpenTime:       "07:00",
			CloseTime:      "22:00",
			Categories:     []string{"Mie", "Snack", "Kopi"},
			Menu: []seedMenu{
				{"Mie", "Mie Ayam Bakso", "Mie ayam topping bakso urat.", 16000},
				{"Mie", "Mie Goreng Pedas", "Mie goreng cabai rawit level 3.", 14000},
				{"Snack", "Pisang Goreng", "Pisang kepok krispi (5 pcs).", 8000},
				{"Snack", "Tahu Crispy", "Tahu goreng tepung krispi.", 9000},
				{"Kopi", "Kopi Susu Gula Aren", "Kopi espresso, susu segar, gula aren.", 15000},
				{"Kopi", "Americano", "Espresso double, air panas.", 12000},
				{"Kopi", "Es Kopi Hitam", "Kopi tubruk dingin, gula opsional.", 10000},
			},
		},
	}

	for _, s := range stores {
		sellerID := userIDByUsername[s.SellerUsername]
		if sellerID == "" {
			log.Fatalf("[seed] seller %s not found", s.SellerUsername)
		}
		storeID, err := upsertStore(db, sellerID, s)
		if err != nil {
			log.Fatalf("[seed] store %s: %v", s.Name, err)
		}
		fmt.Printf("  store %-14s id=%s seller=%s\n", s.Name, storeID, s.SellerUsername)

		catIDByName := map[string]string{}
		for i, name := range s.Categories {
			cid, err := upsertCategory(db, storeID, name, i)
			if err != nil {
				log.Fatalf("[seed] category %s/%s: %v", s.Name, name, err)
			}
			catIDByName[name] = cid
		}
		for _, m := range s.Menu {
			catID := catIDByName[m.Category]
			if catID == "" {
				log.Fatalf("[seed] menu %s references unknown category %s", m.Name, m.Category)
			}
			if err := upsertMenuItem(db, storeID, catID, m); err != nil {
				log.Fatalf("[seed] menu %s: %v", m.Name, err)
			}
		}
		fmt.Printf("        seeded %d categories, %d menu items\n", len(s.Categories), len(s.Menu))
	}

	fmt.Println()
	fmt.Println("=== Seed complete ===")
	fmt.Println("Login credentials (default):")
	fmt.Println("  ADMIN   admin       / admin123")
	fmt.Println("  SELLER  warungbudi  / seller123   (toko: Warung Budi)")
	fmt.Println("  SELLER  kantinmaya  / seller123   (toko: Kantin Maya)")
	fmt.Println("  BUYER   pembeli1    / buyer123")
	fmt.Println("  BUYER   pembeli2    / buyer123")
	fmt.Println("  BUYER   mobileuser  / mobile123   (untuk app Expo)")
}

// upsertUser inserts or updates a user identified by username. Password is
// re-hashed only on insert; existing users keep their current password so
// re-seeding doesn't reset credentials a developer may have changed.
func upsertUser(db *gorm.DB, u seedUser) (string, error) {
	var existing models.User
	err := db.Where("username = ?", u.Username).First(&existing).Error
	if err == nil {
		// Update mutable fields only.
		updates := map[string]any{
			"name":      u.Name,
			"role":      u.Role,
			"email":     u.Email,
			"phone":     u.Phone,
			"is_active": true,
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			return "", err
		}
		return existing.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	hash, err := utils.HashPassword(u.Password)
	if err != nil {
		return "", err
	}
	email := u.Email
	phone := u.Phone
	user := models.User{
		ID:       utils.NewID(),
		Username: u.Username,
		Name:     u.Name,
		Email:    &email,
		Phone:    &phone,
		Password: hash,
		Role:     u.Role,
		IsActive: true,
	}
	if err := db.Create(&user).Error; err != nil {
		return "", err
	}
	return user.ID, nil
}

func upsertStore(db *gorm.DB, sellerID string, s seedStore) (string, error) {
	var existing models.Store
	err := db.Where("seller_id = ?", sellerID).First(&existing).Error
	if err == nil {
		desc := s.Description
		updates := map[string]any{
			"name":        s.Name,
			"description": &desc,
			"address":     s.Address,
			"phone":       s.Phone,
			"open_time":   s.OpenTime,
			"close_time":  s.CloseTime,
			"is_open":     true,
			"is_verified": true,
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			return "", err
		}
		return existing.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return "", err
	}
	desc := s.Description
	store := models.Store{
		ID:          utils.NewID(),
		SellerID:    sellerID,
		Name:        s.Name,
		Description: &desc,
		Address:     s.Address,
		Phone:       s.Phone,
		OpenTime:    s.OpenTime,
		CloseTime:   s.CloseTime,
		IsOpen:      true,
		IsVerified:  true,
	}
	if err := db.Create(&store).Error; err != nil {
		return "", err
	}
	return store.ID, nil
}

func upsertCategory(db *gorm.DB, storeID, name string, sortOrder int) (string, error) {
	var existing models.Category
	err := db.Where("store_id = ? AND name = ?", storeID, name).First(&existing).Error
	if err == nil {
		db.Model(&existing).Update("sort_order", sortOrder)
		return existing.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return "", err
	}
	cat := models.Category{
		ID:        utils.NewID(),
		StoreID:   storeID,
		Name:      name,
		SortOrder: sortOrder,
	}
	if err := db.Create(&cat).Error; err != nil {
		return "", err
	}
	return cat.ID, nil
}

func upsertMenuItem(db *gorm.DB, storeID, categoryID string, m seedMenu) error {
	var existing models.MenuItem
	err := db.Where("store_id = ? AND name = ?", storeID, m.Name).First(&existing).Error
	if err == nil {
		desc := m.Desc
		updates := map[string]any{
			"category_id":  &categoryID,
			"description":  &desc,
			"price":        m.Price,
			"is_available": true,
		}
		return db.Model(&existing).Updates(updates).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	desc := m.Desc
	item := models.MenuItem{
		ID:          utils.NewID(),
		StoreID:     storeID,
		CategoryID:  &categoryID,
		Name:        m.Name,
		Description: &desc,
		Price:       m.Price,
		IsAvailable: true,
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&item).Error
}
