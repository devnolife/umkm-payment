package dto

type RegisterInput struct {
	Username string  `json:"username" validate:"required,min=3,max=32"`
	Password string  `json:"password" validate:"required,min=6"`
	Name     string  `json:"name" validate:"required,min=1"`
	Email    *string `json:"email" validate:"omitempty,email"`
	Phone    *string `json:"phone"`
	Role     string  `json:"role" validate:"omitempty,oneof=BUYER SELLER"`
}

type LoginInput struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileInput struct {
	Name   *string `json:"name"`
	Email  *string `json:"email" validate:"omitempty,email"`
	Phone  *string `json:"phone"`
	Avatar *string `json:"avatar"`
}

type CreateStoreInput struct {
	Name        string   `json:"name" validate:"required"`
	Description *string  `json:"description"`
	Address     string   `json:"address" validate:"required"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	Phone       string   `json:"phone" validate:"required"`
	Image       *string  `json:"image"`
	OpenTime    *string  `json:"openTime"`
	CloseTime   *string  `json:"closeTime"`
}

type UpdateStoreInput struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Address     *string  `json:"address"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	Phone       *string  `json:"phone"`
	Image       *string  `json:"image"`
	OpenTime    *string  `json:"openTime"`
	CloseTime   *string  `json:"closeTime"`
	IsOpen      *bool    `json:"isOpen"`
}

type CategoryInput struct {
	Name      string `json:"name" validate:"required"`
	SortOrder *int   `json:"sortOrder"`
	StoreID   string `json:"storeId" validate:"required"`
}

type MenuItemInput struct {
	StoreID     string  `json:"storeId" validate:"required"`
	CategoryID  *string `json:"categoryId"`
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
	Price       int     `json:"price" validate:"required,min=0"`
	Image       *string `json:"image"`
	IsAvailable *bool   `json:"isAvailable"`
}

type UpdateMenuItemInput struct {
	CategoryID  *string `json:"categoryId"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int    `json:"price" validate:"omitempty,min=0"`
	Image       *string `json:"image"`
	IsAvailable *bool   `json:"isAvailable"`
}

type CreateOrderItemInput struct {
	MenuItemID string  `json:"menuItemId" validate:"required"`
	Quantity   int     `json:"quantity" validate:"required,min=1"`
	Notes      *string `json:"notes"`
}

type CreateOrderInput struct {
	StoreID       string                 `json:"storeId" validate:"required"`
	Items         []CreateOrderItemInput `json:"items" validate:"required,min=1,dive"`
	PaymentMethod string                 `json:"paymentMethod" validate:"required,oneof=COD ONLINE"`
	Notes         *string                `json:"notes"`
}

type UpdateOrderStatusInput struct {
	Status string `json:"status" validate:"required,oneof=PENDING CONFIRMED PROCESSING READY COMPLETED CANCELLED"`
}

type CreatePaymentInput struct {
	OrderID string `json:"orderId" validate:"required"`
}
