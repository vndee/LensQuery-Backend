package main

import (
	"context"
	"flag"
	"log"

	firebase "firebase.google.com/go"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	gofiberfirebaseauth "github.com/sacsand/gofiber-firebaseauth"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/handler"
)

var (
	name = flag.String("name", "LensQuery Backend", "Name of the application")
	port = flag.String("port", "3000", "Port to listen on")
	prod = flag.Bool("prod", false, "Enable prefork in Production")
)

func Setup() *fiber.App {
	app := fiber.New(
		fiber.Config{
			AppName:     *name,
			Prefork:     *prod,
			JSONEncoder: sonic.Marshal,
			JSONDecoder: sonic.Unmarshal,
		},
	)

	cleanup, err := database.GetCloudSQLDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer cleanup()

	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Initialize the firebase app.
	fireApp, _ := firebase.NewApp(context.Background(), nil)

	// Check if the firebase app is initialized properly.
	_, err = fireApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(gofiberfirebaseauth.New(gofiberfirebaseauth.Config{
		FirebaseApp: fireApp,
		IgnoreUrls:  []string{"GET::/terms", "GET::/privacy", "POST::/api/v1/subscription/event_hook"},
	}))

	// Routes
	app.Get("/terms", handler.GetTermsOfUse)
	app.Get("/privacy", handler.GetPrivacyPolicy)

	v1 := app.Group("/api/v1")

	ocr := v1.Group("/ocr")
	ocr.Get("/get_equation_token", handler.GetEquationOCRAppToken)
	ocr.Post("/get_free_text", handler.GetFreeTextContent)
	ocr.Post("/get_document_text", handler.GetDocumentTextContent)
	ocr.Post("/get_equation_text", handler.GetEquationTextContent)

	sub := v1.Group("/subscription")
	sub.Post("/verify_receipt_android", handler.VerifyReceiptAndroid)
	sub.Post("/verify_receipt_ios", handler.VerifyReceiptIOS)
	sub.Get("/get_subscription_plan", handler.GetSubscriptionPlan)
	sub.Get("/get_user_subscription", handler.GetUserSubscription)
	sub.Post("/event_hook", handler.EventHook)

	cre := v1.Group("/credit")
	cre.Get("/get_user_remain_credit", handler.GetUserRemainCredits)
	cre.Get("/do_decrease_credit", handler.DoDecreaseCredits)

	return app
}

func main() {
	flag.Parse()
	log.Println("Starting server...", *name, *port, *prod)

	app := Setup()
	log.Fatal(app.Listen(":" + *port))
}
