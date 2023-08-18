package handler

import (
	"bytes"
	"encoding/json"
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

func Image2Text(c *fiber.Ctx) error {
	log.Println("OCR request received")
	log.Println("OCR_KEY:", MathpixKey)
	log.Println("OCR_APP:", MathpixApp)
	log.Println("OCR_URL:", MathpixURL)

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file"})
	}

	fileReader, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error reading file"})
	}
	defer fileReader.Close()

	imgBytes, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error reading file"})
	}

	req, err := http.NewRequest("POST", MathpixURL, bytes.NewReader(imgBytes))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error creating request"})
	}

	req.Header.Set("app_id", MathpixApp)
	req.Header.Set("app_key", MathpixKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	log.Println("OCR request sent", req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error making request"})
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error reading response"})
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return c.JSON(result)
}
