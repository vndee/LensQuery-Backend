package handler

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/config"
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

	log.Println(resp.Body)

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
