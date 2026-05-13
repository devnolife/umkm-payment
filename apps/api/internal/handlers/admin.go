package handlers

import (
	"time"

	"github.com/devnolife/umkm-api/internal/database"
	"github.com/devnolife/umkm-api/internal/models"
	"github.com/devnolife/umkm-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func AdminStats(c *fiber.Ctx) error {
	var (
		totalUsers   int64
		totalStores  int64
		totalOrders  int64
		todayOrders  int64
		todayRevenue int64
	)
	database.DB.Model(&models.User{}).Count(&totalUsers)
	database.DB.Model(&models.Store{}).Count(&totalStores)
	database.DB.Model(&models.Order{}).Count(&totalOrders)

	start := time.Now().Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	database.DB.Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&todayOrders)
	database.DB.Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ? AND payment_status = ?", start, end, models.PayPaid).
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&todayRevenue)

	var recent []models.Order
	database.DB.
		Preload("Buyer").
		Preload("Store").
		Preload("OrderItems.MenuItem").
		Order("created_at desc").
		Limit(10).
		Find(&recent)

	return utils.OK(c, fiber.Map{
		"totalUsers":   totalUsers,
		"totalStores":  totalStores,
		"totalOrders":  totalOrders,
		"todayOrders":  todayOrders,
		"todayRevenue": todayRevenue,
		"recentOrders": recent,
	})
}

func AdminListUsers(c *fiber.Ctx) error {
	role := c.Query("role")
	q := c.Query("q")
	if q == "" {
		q = c.Query("search")
	}

	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := c.QueryInt("limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}

	tx := database.DB.Model(&models.User{}).Order("created_at desc")
	if role != "" {
		tx = tx.Where("role = ?", role)
	}
	if q != "" {
		tx = tx.Where("LOWER(name) LIKE ? OR LOWER(username) LIKE ?", "%"+q+"%", "%"+q+"%")
	}

	var total int64
	tx.Count(&total)

	var users []models.User
	if err := tx.Offset((page - 1) * limit).Limit(limit).Find(&users).Error; err != nil {
		return utils.Internal(c, "db error")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}
	return utils.OKPaginated(c, users, fiber.Map{
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	})
}

func AdminToggleUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "user not found")
	}
	user.IsActive = !user.IsActive
	database.DB.Save(&user)
	return utils.OK(c, user)
}
