package main

import (
	"flag"
	"log"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	gofiberfirebaseauth "github.com/sacsand/gofiber-firebaseauth"
	"github.com/vndee/lensquery-backend/pkg/config"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/handler"
	"github.com/vndee/lensquery-backend/pkg/limiter"
	"github.com/vndee/lensquery-backend/pkg/templates"
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

	config.SetupFirebase()
	err := database.ConnectRedis()
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	err = limiter.InitLimter()
	if err != nil {
		log.Fatalf("Failed to init limiter: %v", err)
	}

	cleanup, err := database.GetCloudSQLDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer cleanup()

	database.CreateTables()

	err = config.LoadSubscriptionPlanConfig()
	if err != nil {
		log.Fatalf("Failed to load subscription plan config: %v", err)
	}

	err = templates.Load()
	if err != nil {
		log.Fatalf("Failed to load email templates: %v", err)
	}

	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(gofiberfirebaseauth.New(gofiberfirebaseauth.Config{
		FirebaseApp: config.FirebaseApp,
		IgnoreUrls: []string{
			"GET::/terms",
			"GET::/privacy",
			"GET::/api/v1/email/send",
			"POST::/api/v1/subscription/event_hook",
			"GET::/api/v1/account/reset_password",
			"GET::/api/v1/account/verify_code",
		}}))

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
	sub.Post("/event_hook", handler.EventHook)

	cre := v1.Group("/credit")
	cre.Get("/details", handler.GetUserRemainCredits)

	acc := v1.Group("/account")
	acc.Get("/reset_password", handler.RequestResetPasswordCode)
	acc.Get("/verify_code", handler.VerifyCode)

	return app
}

func main() {
	flag.Parse()
	log.Println("Starting server...", *name, *port, *prod)

	app := Setup()
	log.Fatal(app.Listen(":" + *port))
}
