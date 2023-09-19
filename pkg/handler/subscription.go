package handler

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/config"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
)

func handleTestEvent(event *model.Event) {

}

func handleInitialPurchaseEvent(event *model.Event) *model.UserCredits {
	log.Println("Purchased plan:", config.PlanConfigs[event.ProductID])

	userCredits := model.UserCredits{
		UserID:               event.AppUserID,
		PurchasedTimestampMs: event.PurchasedAtMs,
		ExpiredTimestampMs:   event.ExpirationAtMs,
		AmmountEquationSnap:  config.PlanConfigs[event.ProductID].EquationOCRSnap,
		RemainEquationSnap:   config.PlanConfigs[event.ProductID].EquationOCRSnap,
		AmmountTextSnap:      config.PlanConfigs[event.ProductID].TextOCRSnap,
		RemainTextSnap:       config.PlanConfigs[event.ProductID].TextOCRSnap,
	}

	if database.Pool.Where("user_id = ?", event.AppUserID).First(&userCredits).RowsAffected == 0 {
		database.Pool.Create(&userCredits)
	} else {
		database.Pool.Model(&model.UserCredits{}).Where("user_id = ?", event.AppUserID).Updates(userCredits)
	}

	return &userCredits
}

func EventHook(c *fiber.Ctx) error {
	// Check API Bearer token in the header
	if c.Get("Authorization") != "Bearer "+os.Getenv("WEBHOOK_BEARER") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get the JSON body from request
	payload := new(model.WebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	event := payload.Event
	switch event.Type {
	case "TEST":
		handleTestEvent(&event)

	case "INITIAL_PURCHASE":
		response := handleInitialPurchaseEvent(&event)
		return c.Status(fiber.StatusOK).JSON(response)

	case "RENEWAL":
		break

	case "CANCELLATION":
		break

	case "UNCANCELLATION":
		break

	case "NON_RENEWING_PURCHASE":
		break

	case "SUBSCRIPTION_PAUSED":
		break

	case "EXPIRATION":
		break

	case "BILLING_ISSUE":
		break

	case "PRODUCT_CHANGE":
		break

	default:
		break
	}

	return c.Status(fiber.StatusOK).JSON(payload)
}
