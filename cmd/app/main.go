package main

import (
	"context"
	"flag"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	gofiberfirebaseauth "github.com/sacsand/gofiber-firebaseauth"
	"github.com/vndee/lensquery-backend/pkg/handler"
	"google.golang.org/api/option"
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

	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Get google service account credentials
	serviceAccount, fileExi := os.LookupEnv("SERVICE_ACCOUNT_JSON")
	if !fileExi {
		log.Fatal("Please provide valid firebbase auth credential json!", serviceAccount, fileExi)
	}

	// Initialize the firebase app.
	opt := option.WithCredentialsFile(serviceAccount)
	fireApp, _ := firebase.NewApp(context.Background(), nil, opt)

	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(gofiberfirebaseauth.New(gofiberfirebaseauth.Config{
		FirebaseApp: fireApp,
		IgnoreUrls:  []string{},
	}))

	// Routes
	v1 := app.Group("/api/v1")

	ocr := v1.Group("/ocr")
	ocr.Get("/get_access_token", handler.GetAppToken)

	log.Fatal(app.Listen(":" + *port))
}
