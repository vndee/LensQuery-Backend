package handler

import (
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"github.com/gofiber/fiber/v2"
)

var (
	MathpixKey = os.Getenv("OCR_KEY")
	MathpixApp = os.Getenv("OCR_APP")
	MathpixURL = os.Getenv("OCR_URL")
)

// Get short-lived access token
func GetAppToken(c *fiber.Ctx) error {
	req, err := http.NewRequest("POST", MathpixURL+"/app-tokens", nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create request")
	}

	req.Header.Set("app_key", MathpixKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to send request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to read response body")
	}

	// Add app_id to the body
	body = append(body, []byte(`{"app_id": "`+MathpixApp+`"}`)...)
	return c.Status(resp.StatusCode).Send(body)
}
