package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
)

func VerifyReceiptAndroid(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func VerifyReceiptIOS(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func GetSubscriptionPlan(c *fiber.Ctx) error {
	subcriptions := []model.SubcriptionPlan{}
	database.Pool.Find(&subcriptions)

	return c.Status(fiber.StatusOK).JSON(subcriptions)
}

func GetUserSubscription(c *fiber.Ctx) error {
	subcription := model.UserSubscription{}
	database.Pool.Where("user_id = ?", c.Locals("user_id")).First(&subcription)

	return c.Status(fiber.StatusOK).JSON(subcription)
}
