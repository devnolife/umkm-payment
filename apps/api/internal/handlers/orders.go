package handlers

import (
	"time"

	"github.com/devnolife/umkm-api/internal/database"
	"github.com/devnolife/umkm-api/internal/dto"
	"github.com/devnolife/umkm-api/internal/middleware"
	"github.com/devnolife/umkm-api/internal/models"
	"github.com/devnolife/umkm-api/internal/utils"
	"github.com/devnolife/umkm-api/internal/ws"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateOrder(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var in dto.CreateOrderInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	// Fetch menu items in a single query to compute totals from server-side prices.
	itemIDs := make([]string, 0, len(in.Items))
	qtyByID := make(map[string]int, len(in.Items))
	notesByID := make(map[string]*string, len(in.Items))
	for _, it := range in.Items {
		itemIDs = append(itemIDs, it.MenuItemID)
		qtyByID[it.MenuItemID] = it.Quantity
		notesByID[it.MenuItemID] = it.Notes
	}

	var menuItems []models.MenuItem
	if err := database.DB.Where("id IN ?", itemIDs).Find(&menuItems).Error; err != nil {
		return utils.Internal(c, "db error")
	}
	if len(menuItems) != len(itemIDs) {
		return utils.BadRequest(c, "some menu items not found")
	}

	total := 0
	orderItems := make([]models.OrderItem, 0, len(menuItems))
	for _, mi := range menuItems {
		if mi.StoreID != in.StoreID {
			return utils.BadRequest(c, "menu item does not belong to the store")
		}
		if !mi.IsAvailable {
			return utils.BadRequest(c, "menu item not available: "+mi.Name)
		}
		q := qtyByID[mi.ID]
		total += mi.Price * q
		orderItems = append(orderItems, models.OrderItem{
			ID:         utils.NewID(),
			MenuItemID: mi.ID,
			Quantity:   q,
			Price:      mi.Price,
			Notes:      notesByID[mi.ID],
		})
	}

	order := models.Order{
		ID:            utils.NewID(),
		OrderNumber:   utils.GenerateOrderNumber(time.Now().Format("20060102")),
		BuyerID:       userID,
		StoreID:       in.StoreID,
		Status:        models.OrderPending,
		TotalPrice:    total,
		PaymentMethod: models.PaymentMethod(in.PaymentMethod),
		PaymentStatus: models.PayUnpaid,
		Notes:         in.Notes,
	}
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return utils.Internal(c, "failed to create order")
	}

	var full models.Order
	database.DB.Preload("OrderItems.MenuItem").Preload("Buyer").First(&full, "id = ?", order.ID)

	// Realtime notify store/seller.
	ws.Default().Emit("store:"+order.StoreID, "order.created", full)
	ws.Default().Emit("user:"+userID, "order.created", full)

	return utils.Created(c, full)
}

func ListOrders(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)
	status := c.Query("status")
	storeID := c.Query("storeId")

	tx := database.DB.Model(&models.Order{}).
		Preload("OrderItems.MenuItem").
		Preload("Store").
		Preload("Buyer").
		Order("created_at desc")

	switch role {
	case models.RoleBuyer:
		tx = tx.Where("buyer_id = ?", userID)
	case models.RoleSeller:
		// orders for stores owned by this seller
		var store models.Store
		if err := database.DB.Select("id").Where("seller_id = ?", userID).First(&store).Error; err != nil {
			return utils.OK(c, []models.Order{})
		}
		tx = tx.Where("store_id = ?", store.ID)
	}

	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	if storeID != "" && role == models.RoleAdmin {
		tx = tx.Where("store_id = ?", storeID)
	}

	var orders []models.Order
	if err := tx.Find(&orders).Error; err != nil {
		return utils.Internal(c, "db error")
	}
	return utils.OK(c, orders)
}

func GetOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var order models.Order
	if err := database.DB.
		Preload("OrderItems.MenuItem").
		Preload("Store").
		Preload("Buyer").
		Preload("Payment").
		First(&order, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "order not found")
	}

	if role == models.RoleBuyer && order.BuyerID != userID {
		return utils.Forbidden(c, "not your order")
	}
	if role == models.RoleSeller {
		if !sellerOwnsStore(userID, order.StoreID) {
			return utils.Forbidden(c, "not your order")
		}
	}
	return utils.OK(c, order)
}

func UpdateOrderStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	var in dto.UpdateOrderStatusInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	var order models.Order
	if err := database.DB.First(&order, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "order not found")
	}

	newStatus := models.OrderStatus(in.Status)

	// Authorization rules:
	// - Buyer can only CANCEL their own PENDING order
	// - Seller can transition status for their own stores
	// - Admin can do anything
	switch role {
	case models.RoleBuyer:
		if order.BuyerID != userID {
			return utils.Forbidden(c, "not your order")
		}
		if newStatus != models.OrderCancelled || order.Status != models.OrderPending {
			return utils.Forbidden(c, "buyer can only cancel pending order")
		}
	case models.RoleSeller:
		if !sellerOwnsStore(userID, order.StoreID) {
			return utils.Forbidden(c, "not your store")
		}
	}

	order.Status = newStatus
	if newStatus == models.OrderReady {
		t := time.Now()
		order.EstimatedReadyTime = &t
	}
	database.DB.Save(&order)

	ws.Default().Emit("order:"+order.ID, "order.status", order)
	ws.Default().Emit("user:"+order.BuyerID, "order.status", order)
	ws.Default().Emit("store:"+order.StoreID, "order.status", order)

	return utils.OK(c, order)
}
