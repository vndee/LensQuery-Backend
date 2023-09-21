package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/email"
	"github.com/vndee/lensquery-backend/pkg/model"
)

func SendEmail(c *fiber.Ctx) error {
	recipient := c.Query("recipient")
	if recipient == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	go func() {
		err := email.Send("INITIAL_PURCHASE", recipient, model.EmailData{
			SubscriptionPlan: "Premium Plan",
			TransactionID:    "1234567890",
			PurchaseTime:     "2021-01-01 00:00:00",
			ExpirationTime:   "2021-02-01 00:00:00",
			Price:            "100",
		})

		if err != nil {
			log.Println(err)
			// TODO: Handle sending error
		} else {
			log.Println("Email sent!")
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}
