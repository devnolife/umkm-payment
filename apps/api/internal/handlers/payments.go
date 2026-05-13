package handlers

import (
	"time"

	"github.com/devnolife/umkm-api/internal/database"
	"github.com/devnolife/umkm-api/internal/dto"
	"github.com/devnolife/umkm-api/internal/middleware"
	"github.com/devnolife/umkm-api/internal/models"
	"github.com/devnolife/umkm-api/internal/services"
	"github.com/devnolife/umkm-api/internal/utils"
	"github.com/devnolife/umkm-api/internal/ws"
	"github.com/gofiber/fiber/v2"
)

func CreatePayment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var in dto.CreatePaymentInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	var order models.Order
	if err := database.DB.Preload("OrderItems.MenuItem").Preload("Buyer").
		First(&order, "id = ?", in.OrderID).Error; err != nil {
		return utils.NotFound(c, "order not found")
	}
	if order.BuyerID != userID {
		return utils.Forbidden(c, "not your order")
	}
	if order.PaymentMethod != models.PaymentOnline {
		return utils.BadRequest(c, "order is not online payment")
	}

	items := make([]services.SnapItem, 0, len(order.OrderItems))
	for _, oi := range order.OrderItems {
		name := ""
		if oi.MenuItem != nil {
			name = oi.MenuItem.Name
		}
		items = append(items, services.SnapItem{
			ID:       oi.MenuItemID,
			Name:     name,
			Price:    oi.Price,
			Quantity: oi.Quantity,
		})
	}
	customer := services.SnapCustomer{FirstName: order.Buyer.Name}
	if order.Buyer.Email != nil {
		customer.Email = *order.Buyer.Email
	}
	if order.Buyer.Phone != nil {
		customer.Phone = *order.Buyer.Phone
	}

	snap, err := services.CreateSnapTransaction(order.OrderNumber, order.TotalPrice, items, customer)
	if err != nil {
		return utils.Internal(c, "midtrans: "+err.Error())
	}

	payment := models.Payment{
		ID:                  utils.NewID(),
		OrderID:             order.ID,
		Method:              "snap",
		Amount:              order.TotalPrice,
		Status:              models.PayPending,
		MidtransSnapToken:   &snap.Token,
		MidtransRedirectURL: &snap.RedirectURL,
	}
	// Upsert: if payment already exists, update instead.
	var existing models.Payment
	if err := database.DB.Where("order_id = ?", order.ID).First(&existing).Error; err == nil {
		existing.MidtransSnapToken = &snap.Token
		existing.MidtransRedirectURL = &snap.RedirectURL
		existing.Status = models.PayPending
		existing.Amount = order.TotalPrice
		database.DB.Save(&existing)
		payment = existing
	} else {
		if err := database.DB.Create(&payment).Error; err != nil {
			return utils.Internal(c, "failed to create payment")
		}
	}

	database.DB.Model(&order).Update("payment_status", models.PayPending)

	return utils.Created(c, fiber.Map{
		"snapToken":   snap.Token,
		"redirectUrl": snap.RedirectURL,
		"payment":     payment,
	})
}

// PaymentWebhook handles Midtrans notification.
func PaymentWebhook(c *fiber.Ctx) error {
	var body struct {
		TransactionID     string `json:"transaction_id"`
		TransactionStatus string `json:"transaction_status"`
		OrderID           string `json:"order_id"`
		StatusCode        string `json:"status_code"`
		GrossAmount       string `json:"gross_amount"`
		SignatureKey      string `json:"signature_key"`
		FraudStatus       string `json:"fraud_status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.BadRequest(c, "invalid body")
	}

	if !services.VerifyNotificationSignature(body.OrderID, body.StatusCode, body.GrossAmount, body.SignatureKey) {
		return utils.Unauthorized(c, "invalid signature")
	}

	// Find order by OrderNumber (we sent OrderNumber as midtrans order_id).
	var order models.Order
	if err := database.DB.Where("order_number = ?", body.OrderID).First(&order).Error; err != nil {
		return utils.NotFound(c, "order not found")
	}

	var status models.PaymentStatus
	switch body.TransactionStatus {
	case "capture", "settlement":
		status = models.PayPaid
	case "pending":
		status = models.PayPending
	case "deny", "cancel", "expire":
		status = models.PayFailed
	case "refund", "partial_refund":
		status = models.PayRefunded
	default:
		status = models.PayPending
	}

	updates := map[string]any{
		"status":                  status,
		"midtrans_transaction_id": body.TransactionID,
	}
	if status == models.PayPaid {
		now := time.Now()
		updates["paid_at"] = now
	}
	database.DB.Model(&models.Payment{}).Where("order_id = ?", order.ID).Updates(updates)
	database.DB.Model(&order).Update("payment_status", status)

	if status == models.PayPaid && order.Status == models.OrderPending {
		database.DB.Model(&order).Update("status", models.OrderConfirmed)
		order.Status = models.OrderConfirmed
	}

	ws.Default().Emit("order:"+order.ID, "payment.update", fiber.Map{"status": status})
	ws.Default().Emit("user:"+order.BuyerID, "payment.update", fiber.Map{"orderId": order.ID, "status": status})
	ws.Default().Emit("store:"+order.StoreID, "payment.update", fiber.Map{"orderId": order.ID, "status": status})

	return utils.OK(c, fiber.Map{"received": true})
}
