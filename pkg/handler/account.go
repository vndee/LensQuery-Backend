package handler

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"firebase.google.com/go/auth"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/shareed2k/go_limiter"
	"github.com/vndee/lensquery-backend/pkg/config"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/email"
	"github.com/vndee/lensquery-backend/pkg/limiter"
	"github.com/vndee/lensquery-backend/pkg/model"
)

var limiterEmailConfig *go_limiter.Limit = &go_limiter.Limit{
	Algorithm: go_limiter.SlidingWindowAlgorithm,
	Rate:      100,
	Burst:     1,
	Period:    30 * 60 * time.Second, // period of 30 minutes
}

var limiterIPConfig *go_limiter.Limit = &go_limiter.Limit{
	Algorithm: go_limiter.SlidingWindowAlgorithm,
	Rate:      100,
	Burst:     1,
	Period:    30 * 60 * time.Second, // period of 30 minutes
}

func RequestResetPasswordCode(c *fiber.Ctx) error {
	recipient := c.Query("recipient")
	if recipient == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	res, err := limiter.Limiter.Allow(c.Context(), recipient, limiterEmailConfig)
	if err != nil {
		log.Println("Limiter:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !res.Allowed {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"type": "EMAIL",
		})
	}

	res, err = limiter.Limiter.Allow(c.Context(), c.IP(), limiterIPConfig)
	if err != nil {
		log.Println("Limiter:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !res.Allowed {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"type": "IP",
		})
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

	codeMap := map[string]string{
		"type": "RESET_PASSWORD",
		"code": code,
	}
	codeDict, err := sonic.Marshal(&codeMap)
	if err != nil {
		log.Println("Marshal:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	key := fmt.Sprintf("%s_%s", codeData.Type, codeData.Email)
	exists, err := database.RedisClient.Exists(c.Context(), key).Result()
	if err != nil {
		log.Println("Redis:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if exists > 0 {
		err = database.RedisClient.Set(c.Context(), key, codeDict, 15*time.Minute).Err()
		if err != nil {
			log.Println("Redis:", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	} else {
		set, err := database.RedisClient.SetNX(c.Context(), key, codeDict, 5*time.Minute).Result()
		if err != nil {
			log.Println("Redis:", set, err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if !set {
			log.Println("Redis: SetNX failed", set, err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}
	err = email.SendVerificationCode("RESET_PASSWORD", recipient, codeData)
	if err != nil {
		log.Println("Send email:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	})
}

func VerifyCode(c *fiber.Ctx) error {
	verifyType := c.Query("type")
	email := c.Query("email")
	code := c.Query("code")

	if !(verifyType == "RESET_PASSWORD" || verifyType == "VERIFY_EMAIL") || email == "" || code == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	codeDict, err := database.RedisClient.Get(c.Context(), fmt.Sprintf("%s_%s", verifyType, email)).Result()
	if err != nil {
		log.Println("Redis:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	var codeMap map[string]string
	err = sonic.Unmarshal([]byte(codeDict), &codeMap)
	if err != nil {
		log.Println("Unmarshal:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if codeMap["type"] != verifyType || codeMap["code"] != code {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.SendStatus(fiber.StatusOK)
}

func ResetPassword(c *fiber.Ctx) error {
	// parse params from request body
	params := model.ResetPasswordParams{}
	if err := c.BodyParser(&params); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// validate params
	if params.Code == "" || params.NewPassword == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// check if code is valid
	key := fmt.Sprintf("%s_%s", "RESET_PASSWORD", params.Email)
	codeDict, err := database.RedisClient.Get(c.Context(), key).Result()
	if err != nil {
		log.Println("Redis:", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var codeMap map[string]string
	err = sonic.Unmarshal([]byte(codeDict), &codeMap)
	if err != nil {
		log.Println("Unmarshal:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if codeMap["type"] != "RESET_PASSWORD" || codeMap["code"] != params.Code {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// get user by email
	user, err := config.FirebaseAuth.GetUserByEmail(c.Context(), params.Email)
	if err != nil {
		log.Println("Firebase:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// update password
	updateParams := (&auth.UserToUpdate{}).Password(params.NewPassword)
	_, err = config.FirebaseAuth.UpdateUser(c.Context(), user.UID, updateParams)
	if err != nil {
		log.Println("Firebase:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// delete code
	_ = database.RedisClient.Del(c.Context(), key)
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
