package handler

import (
	"io/ioutil"

	"github.com/gofiber/fiber/v2"
)

func GetTermsOfUse(c *fiber.Ctx) error {
	content, err := ioutil.ReadFile("assets/terms.html")
	if err != nil {
		return err
	}

	return c.SendString(string(content))
}

func GetPrivacyPolicy(c *fiber.Ctx) error {
	content, err := ioutil.ReadFile("assets/privacy.html")
	if err != nil {
		return err
	}

	return c.SendString(string(content))
}
