package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
)

func GetUserRemainCredits(c *fiber.Ctx) error {
	credits := model.UserCredits{}
	database.Pool.Where("user_id = ?", c.Locals("user_id")).First(&credits)

	return c.Status(fiber.StatusOK).JSON(credits)
}
