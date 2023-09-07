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
	"github.com/vndee/lensquery-backend/pkg/handler"
	"github.com/vndee/lensquery-backend/pkg/repository"
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

	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/terms", handler.GetTermsOfUse)
	app.Get("/privacy", handler.GetPrivacyPolicy)

	// Get google service account credentials
	// serviceAccount, fileExi := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	// if !fileExi {
	// 	log.Fatal("Please provide valid firebase auth credential json!")
	// 	log.Fatal("serviceAccount:", serviceAccount)
	// 	log.Fatal("fileExi:", fileExi)
	// }

	// Initialize the firebase app.
	// opt := option.WithCredentialsFile(serviceAccount)
	fireApp, _ := firebase.NewApp(context.Background(), nil)

	// Initialize the Google Cloud Vision client.
	err := repository.GCVClient.Init()
	if err != nil {
		log.Fatal(err)
	}

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
	ocr.Get("/get_equation_token", handler.GetEquationOCRAppToken)
	ocr.Post("/get_free_text", handler.GetFreeTextContent)
	ocr.Post("/get_document_text", handler.GetDocumentTextContent)
	ocr.Post("/get_equation_text", handler.GetEquationTextContent)

	sub := v1.Group("/subscription")
	sub.Post("/verify_receipt_android", handler.VerifyReceiptAndroid)
	sub.Post("/verify_receipt_ios", handler.VerifyReceiptIOS)
	sub.Get("/get_subscription_plan", handler.GetSubscriptionPlan)
	sub.Get("/get_user_subscription", handler.GetUserSubscription)
	sub.Get("/get_user_remain_snap", handler.GetUserRemainSnap)
	sub.Get("/do_decrease_snap_credit", handler.DoDecreaseSnapCredit)

	return app
}

func main() {
	flag.Parse()
	log.Println("Starting server...", *name, *port, *prod)

	app := Setup()
	log.Fatal(app.Listen(":" + *port))
}
