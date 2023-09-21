package handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/config"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/email"
	"github.com/vndee/lensquery-backend/pkg/model"
	"gorm.io/gorm"
)

func handleTestEvent(event *model.Event) {

}

func handleInitialPurchaseEvent(event *model.Event) (*model.UserCredits, error) {
	plan := config.PlanConfigs[event.ProductID]

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

	sendEmail(event.Type, event.AppUserID, model.EmailData{
		SubscriptionPlan: plan.Name,
		TransactionID:    event.TransactionID,
		PurchaseTime:     time.Unix(event.PurchasedAtMs/1000, 0).Format("2006-01-02 15:04:05"),
		ExpirationTime:   time.Unix(event.ExpirationAtMs/1000, 0).Format("2006-01-02 15:04:05"),
		Price:            fmt.Sprintf("%.2f %s", event.PriceInPurchasedCurrency, event.Currency),
	})
	return &userCredits, database.ProcessDatabaseResponse(response)
}

func handleExpirationEvent(event *model.Event) (*model.UserCredits, error) {
	plan := config.PlanConfigs[event.ProductID]

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
		response = database.Pool.Model(&model.UserCredits{}).Where("user_id = ?", event.AppUserID).Updates(map[string]interface{}{
			"expired_timestamp_ms":  event.ExpirationAtMs,
			"ammount_equation_snap": 0,
			"remain_equation_snap":  0,
			"ammount_text_snap":     0,
			"remain_text_snap":      0,
		})
	}

	sendEmail(event.Type, event.AppUserID, model.EmailData{
		SubscriptionPlan: plan.Name,
		TransactionID:    event.TransactionID,
		PurchaseTime:     time.Unix(event.PurchasedAtMs/1000, 0).Format("2006-01-02 15:04:05"),
		ExpirationTime:   time.Unix(event.ExpirationAtMs/1000, 0).Format("2006-01-02 15:04:05"),
		Price:            fmt.Sprintf("%.2f %s", event.PriceInPurchasedCurrency, event.Currency),
	})
	return &userCredits, database.ProcessDatabaseResponse(response)
}

func handleRenewalEvent(event *model.Event) (*model.UserCredits, error) {
	plan := config.PlanConfigs[event.ProductID]

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

	sendEmail(event.Type, event.AppUserID, model.EmailData{
		SubscriptionPlan: plan.Name,
		TransactionID:    event.TransactionID,
		PurchaseTime:     time.Unix(event.PurchasedAtMs/1000, 0).Format("2006-01-02 15:04:05"),
		ExpirationTime:   time.Unix(event.ExpirationAtMs/1000, 0).Format("2006-01-02 15:04:05"),
		Price:            fmt.Sprintf("%.2f %s", event.PriceInPurchasedCurrency, event.Currency),
	})
	return &userCredits, database.ProcessDatabaseResponse(response)
}

func handleCancelationEvent(event *model.Event) (*model.UserCredits, error) {
	plan := config.PlanConfigs[event.ProductID]

	sendEmail(event.Type, event.AppUserID, model.EmailData{
		SubscriptionPlan: plan.Name,
		TransactionID:    event.TransactionID,
		PurchaseTime:     time.Unix(event.PurchasedAtMs/1000, 0).Format("2006-01-02 15:04:05"),
		ExpirationTime:   time.Unix(event.ExpirationAtMs/1000, 0).Format("2006-01-02 15:04:05"),
		Price:            fmt.Sprintf("%.2f %s", event.PriceInPurchasedCurrency, event.Currency),
	})
	return nil, nil
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
		response, err = handleCancelationEvent(&event)

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

func sendEmail(emailType string, recipient string, data model.EmailData) error {
	user, err := config.FirebaseAuth.GetUser(context.Background(), recipient)
	if err != nil {
		log.Println("[Err]", err)
		return err
	}

	go func() {
		err := email.Send(emailType, user.Email, data)

		if err != nil {
			log.Println("[Send email err]", err)
			// TODO: Handle sending error
		} else {
			log.Println("Email sent!")
		}
	}()

	return nil
}
