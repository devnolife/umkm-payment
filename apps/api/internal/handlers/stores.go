package handlers

import (
	"github.com/devnolife/umkm-api/internal/database"
	"github.com/devnolife/umkm-api/internal/dto"
	"github.com/devnolife/umkm-api/internal/middleware"
	"github.com/devnolife/umkm-api/internal/models"
	"github.com/devnolife/umkm-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func ListStores(c *fiber.Ctx) error {
	q := c.Query("q")
	if q == "" {
		q = c.Query("search")
	}
	isOpen := c.Query("isOpen")

	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := c.QueryInt("limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}

	tx := database.DB.Model(&models.Store{})
	if q != "" {
		tx = tx.Where("LOWER(name) LIKE ?", "%"+q+"%")
	}
	if isOpen == "true" {
		tx = tx.Where("is_open = ?", true)
	}

	var total int64
	tx.Count(&total)

	var stores []models.Store
	if err := tx.
		Preload("Seller").
		Order("created_at desc").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&stores).Error; err != nil {
		return utils.Internal(c, "db error")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}
	return utils.OKPaginated(c, stores, fiber.Map{
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	})
}

func GetStore(c *fiber.Ctx) error {
	id := c.Params("id")
	var store models.Store
	if err := database.DB.Preload("Categories").First(&store, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "store not found")
	}
	return utils.OK(c, store)
}

func GetStoreMenu(c *fiber.Ctx) error {
	id := c.Params("id")
	var items []models.MenuItem
	if err := database.DB.Preload("Category").
		Where("store_id = ?", id).
		Order("created_at desc").
		Find(&items).Error; err != nil {
		return utils.Internal(c, "db error")
	}
	return utils.OK(c, items)
}

// GetMyStore returns the authenticated seller's store (or 404 if none).
func GetMyStore(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var store models.Store
	if err := database.DB.
		Preload("Categories").
		Where("seller_id = ?", userID).
		First(&store).Error; err != nil {
		return utils.NotFound(c, "store not found")
	}
	return utils.OK(c, store)
}

func CreateStore(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// Only one store per seller
	var exists models.Store
	if err := database.DB.Where("seller_id = ?", userID).First(&exists).Error; err == nil {
		return utils.BadRequest(c, "seller already has a store")
	}

	var in dto.CreateStoreInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	store := models.Store{
		ID:          utils.NewID(),
		SellerID:    userID,
		Name:        in.Name,
		Description: in.Description,
		Address:     in.Address,
		Latitude:    in.Latitude,
		Longitude:   in.Longitude,
		Phone:       in.Phone,
		Image:       in.Image,
		OpenTime:    "08:00",
		CloseTime:   "21:00",
	}
	if in.OpenTime != nil {
		store.OpenTime = *in.OpenTime
	}
	if in.CloseTime != nil {
		store.CloseTime = *in.CloseTime
	}
	if err := database.DB.Create(&store).Error; err != nil {
		return utils.Internal(c, "failed to create store")
	}
	return utils.Created(c, store)
}

func UpdateStore(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var store models.Store
	if err := database.DB.First(&store, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "store not found")
	}
	if role != models.RoleAdmin && store.SellerID != userID {
		return utils.Forbidden(c, "not your store")
	}

	var in dto.UpdateStoreInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	updates := map[string]any{}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.Address != nil {
		updates["address"] = *in.Address
	}
	if in.Latitude != nil {
		updates["latitude"] = *in.Latitude
	}
	if in.Longitude != nil {
		updates["longitude"] = *in.Longitude
	}
	if in.Phone != nil {
		updates["phone"] = *in.Phone
	}
	if in.Image != nil {
		updates["image"] = *in.Image
	}
	if in.OpenTime != nil {
		updates["open_time"] = *in.OpenTime
	}
	if in.CloseTime != nil {
		updates["close_time"] = *in.CloseTime
	}
	if in.IsOpen != nil {
		updates["is_open"] = *in.IsOpen
	}

	if err := database.DB.Model(&store).Updates(updates).Error; err != nil {
		return utils.Internal(c, "failed to update")
	}
	database.DB.First(&store, "id = ?", id)
	return utils.OK(c, store)
}

func ToggleStore(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var store models.Store
	if err := database.DB.First(&store, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "store not found")
	}
	if role != models.RoleAdmin && store.SellerID != userID {
		return utils.Forbidden(c, "not your store")
	}
	store.IsOpen = !store.IsOpen
	database.DB.Save(&store)
	return utils.OK(c, store)
}

// GetStoreQueue returns position of orderId in store's active queue.
func GetStoreQueue(c *fiber.Ctx) error {
	storeID := c.Params("id")
	orderID := c.Query("orderId")

	var orders []models.Order
	if err := database.DB.
		Where("store_id = ? AND status IN ?", storeID,
			[]models.OrderStatus{models.OrderConfirmed, models.OrderProcessing, models.OrderReady}).
		Order("created_at asc").
		Find(&orders).Error; err != nil {
		return utils.Internal(c, "db error")
	}
	position := -1
	for i, o := range orders {
		if o.ID == orderID {
			position = i + 1
			break
		}
	}
	return utils.OK(c, fiber.Map{"total": len(orders), "position": position})
}
