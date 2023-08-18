package main

import (
	"flag"
	"log"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	name = flag.String("name", "LensQuery Backend", "Name of the application")
	port = flag.String("port", "3000", "Port to listen on")
	prod = flag.Bool("prod", false, "Enable prefork in Production")
)

func main() {
	flag.Parse()
	log.Println("Starting server...", *name, *port, *prod)

	app := fiber.New(
		fiber.Config{
			AppName:     *name,
			Prefork:     *prod,
			JSONEncoder: sonic.Marshal,
			JSONDecoder: sonic.Unmarshal,
		},
	)

	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	log.Fatal(app.Listen(":" + *port))
}
