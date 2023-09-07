package handler

import (
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"github.com/bytedance/sonic"

	"github.com/gofiber/fiber/v2"
)

var (
	MathpixKey = os.Getenv("OCR_KEY")
	MathpixApp = os.Getenv("OCR_APP")
	MathpixURL = os.Getenv("OCR_URL")
)

// Get short-lived access token
func GetEquationOCRAppToken(c *fiber.Ctx) error {
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

	// parse body to json and add field "app_id"
	var data map[string]interface{}
	err = sonic.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Failed to parse response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to parse response body")
	}
	data["app_id"] = MathpixApp
	response, err := sonic.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to marshal response body")
	}

	log.Printf("Successfully get app token: %s", response)

	return c.Status(resp.StatusCode).Send(response)
}

func GetFreeTextContent(c *fiber.Ctx) error {
	return c.SendString("Hello, W orld from FreeText APIs 👋!")
}

func GetDocumentTextContent(c *fiber.Ctx) error {
	return c.SendString("Hello, World from DocumentText APIs 👋!")
}

func GetEquationTextContent(c *fiber.Ctx) error {
	return c.SendString("Hello, World from EquationText APIs 👋!")
}
