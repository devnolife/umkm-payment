package handlers

import (
	"errors"

	"github.com/devnolife/umkm-api/internal/database"
	"github.com/devnolife/umkm-api/internal/dto"
	"github.com/devnolife/umkm-api/internal/middleware"
	"github.com/devnolife/umkm-api/internal/models"
	"github.com/devnolife/umkm-api/internal/services"
	"github.com/devnolife/umkm-api/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Register(c *fiber.Ctx) error {
	var in dto.RegisterInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	role := models.RoleBuyer
	if in.Role == string(models.RoleSeller) {
		role = models.RoleSeller
	}

	// Check uniqueness
	var existing models.User
	if err := database.DB.Where("username = ?", in.Username).First(&existing).Error; err == nil {
		return utils.BadRequest(c, "username already taken")
	}

	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return utils.Internal(c, "failed to hash password")
	}

	user := models.User{
		ID:       utils.NewID(),
		Username: in.Username,
		Name:     in.Name,
		Email:    in.Email,
		Phone:    in.Phone,
		Password: hash,
		Role:     role,
		IsActive: true,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return utils.Internal(c, "failed to create user")
	}

	token, err := services.SignToken(user.ID, user.Role)
	if err != nil {
		return utils.Internal(c, "failed to sign token")
	}

	return utils.Raw(c, fiber.StatusCreated, fiber.Map{"user": user, "token": token})
}

func Login(c *fiber.Ctx) error {
	var in dto.LoginInput
	if err := c.BodyParser(&in); err != nil {
		return utils.BadRequest(c, "invalid body")
	}
	if err := utils.Validate(in); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	var user models.User
	if err := database.DB.Where("username = ?", in.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.Unauthorized(c, "invalid credentials")
		}
		return utils.Internal(c, "db error")
	}
	if !user.IsActive {
		return utils.Forbidden(c, "account disabled")
	}
	if !utils.CheckPassword(user.Password, in.Password) {
		return utils.Unauthorized(c, "invalid credentials")
	}

	token, err := services.SignToken(user.ID, user.Role)
	if err != nil {
		return utils.Internal(c, "failed to sign token")
	}
	return utils.Raw(c, fiber.StatusOK, fiber.Map{"user": user, "token": token})
}

// Refresh issues a new token for the caller using their current valid Bearer token.
// Behaviour:
//   - Requires JWTAuth middleware (token must still be valid / not expired).
//   - Looks up the user to honour current role + active state (revokes if disabled).
//   - Returns a freshly signed token with a new expiry (JWT_EXPIRES_HOURS).
//
// Clients should call this endpoint before the current token expires to perform
// silent token rotation; once a token has expired it can no longer be refreshed.
func Refresh(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return utils.Unauthorized(c, "missing user")
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.Unauthorized(c, "user not found")
		}
		return utils.Internal(c, "db error")
	}
	if !user.IsActive {
		return utils.Forbidden(c, "account disabled")
	}

	token, err := services.SignToken(user.ID, user.Role)
	if err != nil {
		return utils.Internal(c, "failed to sign token")
	}
	return utils.OK(c, fiber.Map{"token": token})
}

func Profile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var user models.User
	if err := database.DB.Preload("Store").First(&user, "id = ?", userID).Error; err != nil {
		return utils.NotFound(c, "user not found")
	}
	return utils.OK(c, user)
}

func UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var in dto.UpdateProfileInput
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
	if in.Email != nil {
		updates["email"] = *in.Email
	}
	if in.Phone != nil {
		updates["phone"] = *in.Phone
	}
	if in.Avatar != nil {
		updates["avatar"] = *in.Avatar
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return utils.Internal(c, "failed to update")
	}

	var user models.User
	database.DB.First(&user, "id = ?", userID)
	return utils.OK(c, user)
}
