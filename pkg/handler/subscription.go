package handler

import "github.com/gofiber/fiber/v2"

func VerifyReceiptAndroid(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func VerifyReceiptIOS(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func GetSubscriptionPlan(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func GetUserSubscription(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func GetUserRemainSnap(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func DoDecreaseSnapCredit(c *fiber.Ctx) error {
	return c.SendString("OK")
}
