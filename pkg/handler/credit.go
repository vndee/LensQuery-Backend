package handler

import (
	"github.com/gofiber/fiber/v2"
	gofiberfirebaseauth "github.com/sacsand/gofiber-firebaseauth"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
)

func GetUserRemainCredits(c *fiber.Ctx) error {
	credits := model.UserCredits{}
	user := c.Locals("user").(gofiberfirebaseauth.User)
	response := database.Pool.Where("user_id = ?", user.UserID).First(&credits)

	if err := database.ProcessDatabaseResponse(response); err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.Status(fiber.StatusOK).JSON(credits)
}
