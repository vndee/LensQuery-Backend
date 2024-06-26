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

	err = database.GetCloudSQLDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	database.CreateTables()

	// err = config.LoadSubscriptionPlanConfig()
	err = config.LoadStorePackagesConfig()
	if err != nil {
		log.Fatalf("Failed to load subscription plan config: %v", err)
	}

	config.SetupOpenRouterClient()

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
			"POST::/api/v1/account/reset_password",
			"POST::/api/v1/account/verify_code",
			"POST::/api/v1/account/update_password",
			"GET::/api/v1/chat/models",
			// "POST::/api/v1/chat/completions",
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
	acc.Post("/activate_free_trial", handler.ActivateUserTrial)
	acc.Post("/check_free_trial", handler.CheckTrialPlan)
	acc.Post("/reset_password", handler.RequestResetPasswordCode)
	acc.Post("/verify_code", handler.VerifyCode)
	acc.Post("/update_password", handler.ResetPassword)
	acc.Delete("/", handler.DeleteAccount)

	chat := v1.Group("/chat")
	chat.Get("/models", handler.ListAvailabelModels)
	chat.Post("/completions", handler.Completion)

	return app
}

func main() {
	flag.Parse()
	log.Println("Starting server...", *name, *port, *prod)

	app := Setup()
	log.Fatal(app.Listen(":" + *port))
}
