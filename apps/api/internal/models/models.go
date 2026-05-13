package models

import "time"

type UserRole string

const (
	RoleBuyer  UserRole = "BUYER"
	RoleSeller UserRole = "SELLER"
	RoleAdmin  UserRole = "ADMIN"
)

type OrderStatus string

const (
	OrderPending    OrderStatus = "PENDING"
	OrderConfirmed  OrderStatus = "CONFIRMED"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderReady      OrderStatus = "READY"
	OrderCompleted  OrderStatus = "COMPLETED"
	OrderCancelled  OrderStatus = "CANCELLED"
)

type PaymentMethod string

const (
	PaymentCOD    PaymentMethod = "COD"
	PaymentOnline PaymentMethod = "ONLINE"
)

type PaymentStatus string

const (
	PayUnpaid   PaymentStatus = "UNPAID"
	PayPending  PaymentStatus = "PENDING"
	PayPaid     PaymentStatus = "PAID"
	PayFailed   PaymentStatus = "FAILED"
	PayRefunded PaymentStatus = "REFUNDED"
)

type User struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Name      string    `gorm:"not null" json:"name"`
	Email     *string   `json:"email,omitempty"`
	Phone     *string   `gorm:"uniqueIndex" json:"phone,omitempty"`
	Password  string    `gorm:"not null" json:"-"`
	Role      UserRole  `gorm:"type:text;default:BUYER;not null" json:"role"`
	Avatar    *string   `json:"avatar,omitempty"`
	IsActive  bool      `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Store  *Store  `gorm:"foreignKey:SellerID;references:ID" json:"store,omitempty"`
	Orders []Order `gorm:"foreignKey:BuyerID;references:ID" json:"-"`
}

func (User) TableName() string { return "users" }

type Store struct {
	ID          string   `gorm:"primaryKey;type:text" json:"id"`
	SellerID    string   `gorm:"uniqueIndex;not null" json:"sellerId"`
	Name        string   `gorm:"not null" json:"name"`
	Description *string  `json:"description,omitempty"`
	Address     string   `gorm:"not null" json:"address"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	Phone       string   `gorm:"not null" json:"phone"`
	Image       *string  `json:"image,omitempty"`
	IsOpen      bool     `gorm:"default:false" json:"isOpen"`
	IsVerified  bool     `gorm:"default:false" json:"isVerified"`
	OpenTime    string   `gorm:"default:08:00" json:"openTime"`
	CloseTime   string   `gorm:"default:21:00" json:"closeTime"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Seller     *User      `gorm:"foreignKey:SellerID;references:ID" json:"seller,omitempty"`
	Categories []Category `gorm:"foreignKey:StoreID;references:ID" json:"categories,omitempty"`
	MenuItems  []MenuItem `gorm:"foreignKey:StoreID;references:ID" json:"menuItems,omitempty"`
	Orders     []Order    `gorm:"foreignKey:StoreID;references:ID" json:"-"`
}

func (Store) TableName() string { return "stores" }

type Category struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	StoreID   string    `gorm:"not null;index:idx_store_cat,unique" json:"storeId"`
	Name      string    `gorm:"not null;index:idx_store_cat,unique" json:"name"`
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Store     *Store     `gorm:"foreignKey:StoreID;references:ID" json:"-"`
	MenuItems []MenuItem `gorm:"foreignKey:CategoryID;references:ID" json:"menuItems,omitempty"`
}

func (Category) TableName() string { return "categories" }

type MenuItem struct {
	ID          string  `gorm:"primaryKey;type:text" json:"id"`
	StoreID     string  `gorm:"not null;index" json:"storeId"`
	CategoryID  *string `gorm:"index" json:"categoryId,omitempty"`
	Name        string  `gorm:"not null" json:"name"`
	Description *string `json:"description,omitempty"`
	Price       int     `gorm:"not null" json:"price"`
	Image       *string `json:"image,omitempty"`
	IsAvailable bool    `gorm:"default:true" json:"isAvailable"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Store    *Store    `gorm:"foreignKey:StoreID;references:ID" json:"-"`
	Category *Category `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
}

func (MenuItem) TableName() string { return "menu_items" }

type Order struct {
	ID            string        `gorm:"primaryKey;type:text" json:"id"`
	OrderNumber   string        `gorm:"uniqueIndex;not null" json:"orderNumber"`
	BuyerID       string        `gorm:"not null;index" json:"buyerId"`
	StoreID       string        `gorm:"not null;index" json:"storeId"`
	Status        OrderStatus   `gorm:"type:text;default:PENDING;not null;index" json:"status"`
	TotalPrice    int           `gorm:"not null" json:"totalPrice"`
	PaymentMethod PaymentMethod `gorm:"type:text;default:COD" json:"paymentMethod"`
	PaymentStatus PaymentStatus `gorm:"type:text;default:UNPAID" json:"paymentStatus"`
	Notes         *string       `json:"notes,omitempty"`

	EstimatedReadyTime *time.Time `json:"estimatedReadyTime,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`

	Buyer      *User       `gorm:"foreignKey:BuyerID;references:ID" json:"buyer,omitempty"`
	Store      *Store      `gorm:"foreignKey:StoreID;references:ID" json:"store,omitempty"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID;references:ID" json:"orderItems,omitempty"`
	Payment    *Payment    `gorm:"foreignKey:OrderID;references:ID" json:"payment,omitempty"`
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID         string  `gorm:"primaryKey;type:text" json:"id"`
	OrderID    string  `gorm:"not null;index" json:"orderId"`
	MenuItemID string  `gorm:"not null;index" json:"menuItemId"`
	Quantity   int     `gorm:"not null" json:"quantity"`
	Price      int     `gorm:"not null" json:"price"`
	Notes      *string `json:"notes,omitempty"`

	Order    *Order    `gorm:"foreignKey:OrderID;references:ID" json:"-"`
	MenuItem *MenuItem `gorm:"foreignKey:MenuItemID;references:ID" json:"menuItem,omitempty"`
}

func (OrderItem) TableName() string { return "order_items" }

type Payment struct {
	ID                    string        `gorm:"primaryKey;type:text" json:"id"`
	OrderID               string        `gorm:"uniqueIndex;not null" json:"orderId"`
	Method                string        `gorm:"not null" json:"method"`
	Amount                int           `gorm:"not null" json:"amount"`
	Status                PaymentStatus `gorm:"type:text;default:UNPAID" json:"status"`
	MidtransTransactionID *string       `json:"midtransTransactionId,omitempty"`
	MidtransSnapToken     *string       `json:"midtransSnapToken,omitempty"`
	MidtransRedirectURL   *string       `json:"midtransRedirectUrl,omitempty"`
	PaidAt                *time.Time    `json:"paidAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Order *Order `gorm:"foreignKey:OrderID;references:ID" json:"-"`
}

func (Payment) TableName() string { return "payments" }
