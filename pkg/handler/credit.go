package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
)

func GetUserRemainCredits(c *fiber.Ctx) error {
	credits := model.UserCredits{}
	response := database.Pool.Where("user_id = ?", c.Locals("user_id")).First(&credits)

	if err := database.ProcessDatabaseResponse(response); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(credits)
}
