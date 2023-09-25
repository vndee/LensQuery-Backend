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
	Rate:      config.EmailLimiterRate,
	Burst:     config.EmailLimiterBurst,
	Period:    config.EmailLimiterPeriod,
}

var limiterIPConfig *go_limiter.Limit = &go_limiter.Limit{
	Algorithm: go_limiter.SlidingWindowAlgorithm,
	Rate:      config.IPLimiterRate,
	Burst:     config.IPLimiterBurst,
	Period:    config.IPLimiterPeriod,
}

func RequestResetPasswordCode(c *fiber.Ctx) error {
	params := model.RequestResetPasswordParams{}
	if err := c.BodyParser(&params); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if params.Recipient == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	res, err := limiter.Limiter.Allow(c.Context(), params.Recipient, limiterEmailConfig)
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
		Email: params.Recipient,
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
		err = database.RedisClient.Set(c.Context(), key, codeDict, config.AccountVerificationCodeTTL).Err()
		if err != nil {
			log.Println("Redis:", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	} else {
		set, err := database.RedisClient.SetNX(c.Context(), key, codeDict, config.AccountVerificationCodeTTL).Result()
		if err != nil {
			log.Println("Redis:", set, err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if !set {
			log.Println("Redis: SetNX failed", set, err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}
	err = email.SendVerificationCode("RESET_PASSWORD", params.Recipient, codeData)
	if err != nil {
		log.Println("Send email:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"exp": time.Now().Add(config.AccountVerificationCodeTTL - 3).Unix(),
	})
}

func VerifyCode(c *fiber.Ctx) error {
	params := model.VerifyResetPasswordParams{}
	if err := c.BodyParser(&params); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !(params.Type == "RESET_PASSWORD" || params.Type == "VERIFY_EMAIL") || params.Email == "" || params.Code == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	codeDict, err := database.RedisClient.Get(c.Context(), fmt.Sprintf("%s_%s", params.Type, params.Email)).Result()
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

	if codeMap["type"] != params.Type || codeMap["code"] != params.Code {
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

func DeleteAccount(c *fiber.Ctx) error {
	params := model.DeleteAccountParams{}
	if err := c.BodyParser(&params); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if params.UserId == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	err := config.FirebaseAuth.DeleteUser(c.Context(), params.UserId)
	if err != nil {
		log.Println("Firebase:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// delete user credits
	_ = database.Pool.Where("user_id = ?", params.UserId).Delete(&model.UserCredits{})
	return c.SendStatus(fiber.StatusOK)
}

func ActivateUserTrial(c *fiber.Ctx) error {
	params := model.ActivateUserTrialParams{}
	if err := c.BodyParser(&params); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if params.UserId == "" || params.Email == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// check if user exists
	user, err := config.FirebaseAuth.GetUser(c.Context(), params.UserId)
	if err != nil {
		log.Println("Firebase:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if user.Email != params.Email {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// check if user already has trial
	trialData := model.UserTrialData{}
	err = database.Pool.Where("user_id = ?", params.UserId).First(&trialData).Error
	if err == nil {
		log.Println("User already has trial")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"exp": trialData.ExpiredTimestampMs,
		})
	}

	if trialData.UserID != "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// create trial data
	trialData = model.UserTrialData{
		UserID:             params.UserId,
		Email:              params.Email,
		ExpiredTimestampMs: time.Now().Add(config.TrialPeriod).Unix(),
	}

	err = database.Pool.Create(&trialData).Error
	if err != nil {
		log.Println("Database:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// create user credits
	credits := model.UserCredits{
		UserID:               params.UserId,
		PurchasedTimestampMs: trialData.CreatedAt.Unix(),
		ExpiredTimestampMs:   trialData.ExpiredTimestampMs,
		AmmountEquationSnap:  config.TrialFreeEquationCredits,
		RemainEquationSnap:   config.TrialFreeEquationCredits,
		AmmountTextSnap:      config.TrialFreeTextSnapCredits,
		RemainTextSnap:       config.TrialFreeTextSnapCredits,
	}

	err = database.Pool.Create(&credits).Error
	if err != nil {
		log.Println("Database:", err)
		// delete trial data
		_ = database.Pool.Where("user_id = ?", params.UserId).Delete(&trialData)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"exp": trialData.ExpiredTimestampMs,
	})
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
