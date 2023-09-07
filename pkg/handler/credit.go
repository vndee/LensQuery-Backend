package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
	"gorm.io/gorm"
)

func GetUserRemainCredits(c *fiber.Ctx) error {
	credits := model.UserCredits{}
	database.DB.DB.Where("user_id = ?", c.Locals("user_id")).First(&credits)
	return c.Status(fiber.StatusOK).JSON(credits)
}

func DoDecreaseCredits(c *fiber.Ctx) error {
	// Update remain credits by -1
	database.DB.DB.Model(&model.UserCredits{}).Where("user_id = ?", c.Locals("user_id")).Update("remain_credits", gorm.Expr("remain_credits - ?", 1))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OK",
	})
}
