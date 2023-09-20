package handler

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/config"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
	"gorm.io/gorm"
)

func handleTestEvent(event *model.Event) {

}

func handleInitialPurchaseEvent(event *model.Event) (*model.UserCredits, error) {
	plan := config.PlanConfigs[event.ProductID]
	log.Println("Purchased plan:", plan)

	userCredits := model.UserCredits{
		UserID:               event.AppUserID,
		PurchasedTimestampMs: event.PurchasedAtMs,
		ExpiredTimestampMs:   event.ExpirationAtMs,
		AmmountEquationSnap:  plan.EquationOCRSnap,
		RemainEquationSnap:   plan.EquationOCRSnap,
		AmmountTextSnap:      plan.TextOCRSnap,
		RemainTextSnap:       plan.TextOCRSnap,
	}

	var response *gorm.DB

	if database.Pool.Where("user_id = ?", event.AppUserID).First(&model.UserCredits{}).RowsAffected == 0 {
		response = database.Pool.Create(&userCredits)
	} else {
		response = database.Pool.Model(&model.UserCredits{}).Where("user_id = ?", event.AppUserID).Updates(userCredits)
	}

	return &userCredits, processDatabaseResponse(response)
}

func handleExpirationEvent(event *model.Event) (*model.UserCredits, error) {
	plan := config.PlanConfigs[event.ProductID]
	log.Println("Expired plan:", plan)

	userCredits := model.UserCredits{
		UserID:               event.AppUserID,
		PurchasedTimestampMs: event.PurchasedAtMs,
		ExpiredTimestampMs:   event.ExpirationAtMs,
		AmmountEquationSnap:  0,
		RemainEquationSnap:   0,
		AmmountTextSnap:      0,
		RemainTextSnap:       0,
	}

	var response *gorm.DB

	if database.Pool.Where("user_id = ?", event.AppUserID).First(&model.UserCredits{}).RowsAffected == 0 {
		response = database.Pool.Create(&userCredits)
	} else {
		response = database.Pool.Model(&model.UserCredits{}).Where("user_id = ?", event.AppUserID).Updates(userCredits)
	}

	return &userCredits, processDatabaseResponse(response)
}

func handleRenewalEvent(event *model.Event) (*model.UserCredits, error) {
	plan := config.PlanConfigs[event.ProductID]
	log.Println("Renewed plan:", plan)

	userCredits := model.UserCredits{
		UserID:               event.AppUserID,
		PurchasedTimestampMs: event.PurchasedAtMs,
		ExpiredTimestampMs:   event.ExpirationAtMs,
		AmmountEquationSnap:  plan.EquationOCRSnap,
		RemainEquationSnap:   plan.EquationOCRSnap,
		AmmountTextSnap:      plan.TextOCRSnap,
		RemainTextSnap:       plan.TextOCRSnap,
	}

	log.Println("User credits:", userCredits)

	var response *gorm.DB
	if database.Pool.Where("user_id = ?", event.AppUserID).First(&model.UserCredits{}).RowsAffected == 0 {
		response = database.Pool.Create(&userCredits)
	} else {
		response = database.Pool.Model(&model.UserCredits{}).Where("user_id = ?", event.AppUserID).Updates(userCredits)
	}

	return &userCredits, processDatabaseResponse(response)
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
	var response *model.UserCredits
	var err error

	switch event.Type {
	case "TEST":
		handleTestEvent(&event)

	case "INITIAL_PURCHASE":
		response, err = handleInitialPurchaseEvent(&event)

	case "RENEWAL":
		response, err = handleRenewalEvent(&event)

	case "CANCELLATION":
		break

	case "UNCANCELLATION":
		break

	case "NON_RENEWING_PURCHASE":
		break

	case "SUBSCRIPTION_PAUSED":
		break

	case "EXPIRATION":
		response, err = handleExpirationEvent(&event)

	case "BILLING_ISSUE":
		break

	case "PRODUCT_CHANGE":
		break

	default:
		break
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func processDatabaseResponse(response *gorm.DB) error {
	if response.Error != nil {
		return response.Error
	}

	if response.RowsAffected == 0 {
		return fiber.ErrNotFound
	}

	return nil
}
