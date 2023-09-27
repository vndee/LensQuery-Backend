package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	gofiberfirebaseauth "github.com/sacsand/gofiber-firebaseauth"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
)

func GetUserRemainCredits(c *fiber.Ctx) error {
	credits := model.UserCredits{}
	user := c.Locals("user").(gofiberfirebaseauth.User)
	response := database.Pool.Where("user_id = ?", user.UserID).First(&credits)

	if credits.ExpiredTimestampMs < time.Now().Unix() {
		_ = database.Pool.Model(&model.UserCredits{}).Where("user_id = ?", user.UserID).Updates(map[string]interface{}{
			"expired_timestamp_ms":  credits.ExpiredTimestampMs,
			"ammount_equation_snap": 0,
			"remain_equation_snap":  0,
			"ammount_text_snap":     0,
			"remain_text_snap":      0,
		})

		return c.SendStatus(fiber.StatusNotFound)
	}

	if err := database.ProcessDatabaseResponse(response); err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.Status(fiber.StatusOK).JSON(credits)
}
