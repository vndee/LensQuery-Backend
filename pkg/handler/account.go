package handler

import (
	"crypto/rand"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/email"
	"github.com/vndee/lensquery-backend/pkg/model"
)

func RequestResetPasswordCode(c *fiber.Ctx) error {
	recipient := c.Query("recipient")
	if recipient == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	code, err := generateRandomCode(6)
	if err != nil {
		log.Println("Generate code:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	codeData := model.VerificationCode{
		Type:  "RESET_PASSWORD",
		Email: recipient,
		Code:  code,
	}

	response := database.Pool.Create(&codeData)
	if err = database.ProcessDatabaseResponse(response); err != nil {
		log.Println("Insert database:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	err = email.SendVerificationCode("RESET_PASSWORD", recipient, codeData)
	if err != nil {
		log.Println("Send email:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func VerifyCode(c *fiber.Ctx) error {
	verifyType := c.Query("type")
	email := c.Query("email")
	code := c.Query("code")

	if !(verifyType == "RESET_PASSWORD" || verifyType == "VERIFY_EMAIL") || email == "" || code == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	codeData := model.VerificationCode{}
	response := database.Pool.Where("type = ? AND email = ? AND code = ?", verifyType, email, code).First(&codeData)
	if err := database.ProcessDatabaseResponse(response); err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	_ = database.Pool.Delete(&codeData)

	return c.SendStatus(fiber.StatusOK)
}

func generateRandomCode(length int) (string, error) {
	const charset = "0123456789"
	code := make([]byte, length)

	_, err := rand.Read(code)
	if err != nil {
		return "", err
	}

	for i := range code {
		code[i] = charset[code[i]%byte(len(charset))]
	}

	return string(code), nil
}
