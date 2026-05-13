package handlers

import (
	"github.com/devnolife/umkm-api/internal/database"
	"github.com/devnolife/umkm-api/internal/dto"
	"github.com/devnolife/umkm-api/internal/middleware"
	"github.com/devnolife/umkm-api/internal/models"
	"github.com/devnolife/umkm-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func sellerOwnsStore(userID, storeID string) bool {
	var store models.Store
	if err := database.DB.Select("seller_id").First(&store, "id = ?", storeID).Error; err != nil {
		return false
	}
	return store.SellerID == userID
}

func CreateCategory(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var in dto.CategoryInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if role != models.RoleAdmin && !sellerOwnsStore(userID, in.StoreID) {
		return utils.Forbidden(c, "not your store")
	}

	cat := models.Category{
		ID:      utils.NewID(),
		StoreID: in.StoreID,
		Name:    in.Name,
	}
	if in.SortOrder != nil {
		cat.SortOrder = *in.SortOrder
	}
	if err := database.DB.Create(&cat).Error; err != nil {
		return utils.Internal(c, "failed to create category")
	}
	return utils.Created(c, cat)
}

func UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var cat models.Category
	if err := database.DB.First(&cat, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "category not found")
	}
	if role != models.RoleAdmin && !sellerOwnsStore(userID, cat.StoreID) {
		return utils.Forbidden(c, "not your store")
	}

	var in dto.CategoryInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	cat.Name = in.Name
	if in.SortOrder != nil {
		cat.SortOrder = *in.SortOrder
	}
	database.DB.Save(&cat)
	return utils.OK(c, cat)
}

func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var cat models.Category
	if err := database.DB.First(&cat, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "category not found")
	}
	if role != models.RoleAdmin && !sellerOwnsStore(userID, cat.StoreID) {
		return utils.Forbidden(c, "not your store")
	}
	database.DB.Delete(&cat)
	return utils.OK(c, fiber.Map{"id": id})
}

func CreateMenuItem(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var in dto.MenuItemInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if role != models.RoleAdmin && !sellerOwnsStore(userID, in.StoreID) {
		return utils.Forbidden(c, "not your store")
	}

	item := models.MenuItem{
		ID:          utils.NewID(),
		StoreID:     in.StoreID,
		CategoryID:  in.CategoryID,
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
		Image:       in.Image,
		IsAvailable: true,
	}
	if in.IsAvailable != nil {
		item.IsAvailable = *in.IsAvailable
	}
	if err := database.DB.Create(&item).Error; err != nil {
		return utils.Internal(c, "failed to create menu item")
	}
	return utils.Created(c, item)
}

func UpdateMenuItem(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var item models.MenuItem
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "menu item not found")
	}
	if role != models.RoleAdmin && !sellerOwnsStore(userID, item.StoreID) {
		return utils.Forbidden(c, "not your store")
	}

	var in dto.UpdateMenuItemInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	updates := map[string]any{}
	if in.CategoryID != nil {
		updates["category_id"] = *in.CategoryID
	}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.Price != nil {
		updates["price"] = *in.Price
	}
	if in.Image != nil {
		updates["image"] = *in.Image
	}
	if in.IsAvailable != nil {
		updates["is_available"] = *in.IsAvailable
	}
	database.DB.Model(&item).Updates(updates)
	database.DB.First(&item, "id = ?", id)
	return utils.OK(c, item)
}

func DeleteMenuItem(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var item models.MenuItem
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "menu item not found")
	}
	if role != models.RoleAdmin && !sellerOwnsStore(userID, item.StoreID) {
		return utils.Forbidden(c, "not your store")
	}
	database.DB.Delete(&item)
	return utils.OK(c, fiber.Map{"id": id})
}
