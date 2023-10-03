package handler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	gofiberfirebaseauth "github.com/sacsand/gofiber-firebaseauth"
	"github.com/sashabaranov/go-openai"
	"github.com/valyala/fasthttp"
	"github.com/vndee/lensquery-backend/pkg/config"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
	"gorm.io/gorm"
)

func ListAvailabelModels(c *fiber.Ctx) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.OpenRouterEndpoint+"/models", nil)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	resp, err := client.Do(req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}

	var responseData map[string]interface{}
	err = sonic.Unmarshal(responseBody, &responseData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}

	return c.Status(fiber.StatusOK).JSON(responseData)
}

func Completion(c *fiber.Ctx) error {
	if !checkAvailableCredit(c) {
		return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{
			"error": "Insufficient credit",
		})
	}

	user := c.Locals("user").(gofiberfirebaseauth.User)

	var requestBody *openai.ChatCompletionRequest
	if err := c.BodyParser(&requestBody); err != nil {
		log.Println("BodyParser:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, "keep-alive")
	c.Set(fiber.HeaderAccessControlAllowOrigin, "*")
	c.Set(fiber.HeaderTransferEncoding, "chunked")

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		stream, err := config.OpenRouterClient.CreateChatCompletionStream(context.Background(), *requestBody)
		if err != nil {
			log.Println("CreateChatCompletionStream:", err)
			return
		}
		defer stream.Close()

		var requestID string

		for {
			response, err := stream.Recv()

			if errors.Is(err, io.EOF) {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				err = w.Flush()
				if err != nil {
					fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)
				}
				log.Println("Checkpoint 1", requestID)
				err := decreaseUserCredit(user.UserID, requestID)
				if err != nil {
					fmt.Printf("Error while decreasing user credit: %v. Closing http connection.\n", err)
				}

				stream.Close()
				break
			}

			requestID = response.ID
			fmt.Fprintf(w, "data: %s\n\n", response.Choices[0].Delta.Content)

			err = w.Flush()
			if err != nil {
				// Refreshing page in web browser will establish a new
				// SSE connection, but only (the last) one is alive, so
				// dead connections must be closed here.
				fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)

				err := decreaseUserCredit(user.UserID, requestID)
				if err != nil {
					fmt.Printf("Error while decreasing user credit: %v. Closing http connection.\n", err)
				}
				stream.Close()
				break
			}
		}
	}))

	return nil
}

func checkAvailableCredit(c *fiber.Ctx) bool {
	user := c.Locals("user").(gofiberfirebaseauth.User)
	var response *gorm.DB
	var userCredits model.UserCredits
	response = database.Pool.Where("user_id = ?", user.UserID).First(&userCredits)
	if err := database.ProcessDatabaseResponse(response); err != nil {
		return false
	}

	if userCredits.CreditAmount <= config.MinPrice {
		return false
	}

	return true
}

func decreaseUserCredit(userID string, requestID string) error {
	url := fmt.Sprintf("%s/generation?id=%s", config.OpenRouterEndpoint, requestID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.OpenRouterAPIKey))
	req.Header.Set("HTTP-Referer", "https://lensquery.com/")
	req.Header.Set("X-Title", "LensQuery")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	chatHistory, err := parseChatHistory(responseBody)
	if err != nil {
		return err
	}

	if chatHistory.Usage <= 0 {
		chatHistory.Usage = config.MinPrice
	} else {
		chatHistory.Usage = chatHistory.Usage * config.PriceAdjustFactor
	}

	var response *gorm.DB
	var userCredits model.UserCredits
	response = database.Pool.Where("user_id = ?", userID).First(&userCredits)
	if err := database.ProcessDatabaseResponse(response); err != nil {
		return err
	}

	response = database.Pool.Model(&model.UserCredits{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"credit_amount": userCredits.CreditAmount - chatHistory.Usage,
	})
	if err := database.ProcessDatabaseResponse(response); err != nil {
		return err
	}

	if database.Pool.Where("id = ?", chatHistory.ID).First(&model.Receipt{}).RowsAffected == 0 {
		response = database.Pool.Create(&chatHistory)
	} else {
		response = database.Pool.Model(&model.Receipt{}).Where("id = ?", chatHistory.ID).Updates(chatHistory)
	}

	return database.ProcessDatabaseResponse(response)
}

func parseChatHistory(responseBody []byte) (*model.Receipt, error) {
	parsedObject := &model.ReceiptResponse{}
	err := sonic.Unmarshal(responseBody, parsedObject)
	return &parsedObject.Data, err
}
