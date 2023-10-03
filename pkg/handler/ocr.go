package handler

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"log"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	gofiberfirebaseauth "github.com/sacsand/gofiber-firebaseauth"
	"github.com/vndee/lensquery-backend/pkg/config"
	"github.com/vndee/lensquery-backend/pkg/database"
	"github.com/vndee/lensquery-backend/pkg/model"
)

var (
	MathpixKey = os.Getenv("OCR_KEY")
	MathpixApp = os.Getenv("OCR_APP")
	MathpixURL = os.Getenv("OCR_URL")
)

const NUM_LABELS = 5
const INTERNAL_SERVER_ERROR = "Internal Server Error"

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
	// Check if user has enough credits
	if !checkAvailableSnapCredits(c, "text") {
		log.Printf("User has not enough credits")
		return c.Status(fiber.StatusPaymentRequired).SendString("Not enough credits")
	}

	// Get image from request body
	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("Image is required")
		return c.Status(fiber.StatusBadRequest).SendString("Image is required")
	}

	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Printf("Failed to create gcv client: %v", err)
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
	}
	defer client.Close()

	fileBytes := make([]byte, file.Size)
	f, err := file.Open()
	if err != nil {
		log.Printf("Failed to read image from request body: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to read image from request body")
	}
	defer f.Close()
	f.Read(fileBytes)

	image, err := vision.NewImageFromReader(bytes.NewReader(fileBytes))
	if err != nil {
		log.Printf("Failed to create image from request body: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create image from request body")
	}

	results := make(map[string]interface{})

	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		log.Printf("Failed to detect texts: %v", err)
		results["text"] = ""
	} else {
		if len(annotations) == 0 {
			log.Printf("No text found")
			results["text"] = ""
		} else {
			log.Printf("Found %d text(s)", len(annotations)-1)
			results["text"] = annotations[0].Description
		}
	}

	labels, err := client.DetectLabels(ctx, image, nil, NUM_LABELS)
	if err != nil {
		log.Printf("Failed to detect labels: %v", err)
		results["labels"] = []string{}
	} else {
		if len(labels) == 0 {
			log.Printf("No label found")
			results["labels"] = []string{}
		} else {
			log.Printf("Found %d labels:", len(labels))

			var labelDescription []string
			for _, annotation := range labels {
				labelDescription = append(labelDescription, annotation.Description)
			}

			results["labels"] = labelDescription
		}
	}

	err = doDecreaseSnapCredits(c, "text")
	if err != nil {
		log.Printf("Failed to decrease snap credits: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	return c.Status(fiber.StatusOK).JSON(results)
}

func GetDocumentTextContent(c *fiber.Ctx) error {
	// Check if user has enough credits
	if !checkAvailableSnapCredits(c, "text") {
		log.Printf("User has not enough credits")
		return c.Status(fiber.StatusPaymentRequired).SendString("Not enough credits")
	}

	file, err := c.FormFile("image")
	if err != nil {
		log.Println("Image is required")
		return c.Status(fiber.StatusBadRequest).SendString("Image is required")
	}

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Printf("Failed to create gcv client: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	defer client.Close()

	fileBytes := make([]byte, file.Size)
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	defer f.Close()
	f.Read((fileBytes))

	image, err := vision.NewImageFromReader(bytes.NewReader(fileBytes))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	results := make(map[string]interface{})

	annotation, err := client.DetectDocumentText(ctx, image, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	} else {
		if annotation == nil {
			log.Println("No text found")
			results["text"] = ""
		} else {
			log.Println("Found text")
			results["text"] = annotation.Text
		}
	}

	labels, err := client.DetectLabels(ctx, image, nil, NUM_LABELS)
	if err != nil {
		log.Printf("Failed to detect labels: %v", err)
		results["labels"] = []string{}
	} else {
		if len(labels) == 0 {
			log.Printf("No label found")
			results["labels"] = []string{}
		} else {
			log.Printf("Found %d labels:", len(labels))

			var labelDescription []string
			for _, annotation := range labels {
				labelDescription = append(labelDescription, annotation.Description)
			}

			results["labels"] = labelDescription
		}
	}

	err = doDecreaseSnapCredits(c, "text")
	if err != nil {
		log.Printf("Failed to decrease snap credits: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return c.Status(fiber.StatusOK).JSON(results)
}

func GetEquationTextContent(c *fiber.Ctx) error {
	// Check if user has enough credits
	if !checkAvailableSnapCredits(c, "equation") {
		log.Printf("User has not enough credits")
		return c.Status(fiber.StatusPaymentRequired).SendString("Not enough credits")
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Image is required")
	}

	imgFile, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}
	defer imgFile.Close()

	imgBytes := make([]byte, file.Size)
	_, err = imgFile.Read(imgBytes)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}

	// Create the options JSON
	options := model.MathpixOptions{
		MathInlineDelimiters: []string{"$", "$"},
		RmSpaces:             true,
	}
	optionsJSON, err := sonic.Marshal(options)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}
	part.Write(imgBytes)

	err = writer.WriteField("options_json", string(optionsJSON))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}

	err = writer.Close()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}

	// Set up the request to Mathpix API
	req, err := http.NewRequest("POST", MathpixURL+"/text", body)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}
	req.Header.Set("app_id", MathpixApp)
	req.Header.Set("app_key", MathpixKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}

	var responseData map[string]interface{}
	err = sonic.Unmarshal(responseBody, &responseData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(INTERNAL_SERVER_ERROR)
	}

	err = doDecreaseSnapCredits(c, "equation")
	if err != nil {
		log.Printf("Failed to decrease snap credits: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	return c.Status(response.StatusCode).JSON(responseData)
}

func checkAvailableSnapCredits(c *fiber.Ctx, snapType string) bool {
	user := c.Locals("user").(gofiberfirebaseauth.User)

	var userCredits model.UserCredits
	response := database.Pool.Where("user_id = ?", user.UserID).First(&userCredits)
	if err := database.ProcessDatabaseResponse(response); err != nil {
		return false
	}

	switch snapType {
	case "equation":
		if userCredits.CreditAmount < config.EquationTextSnapPrice {
			return false
		}

	case "text":
		if userCredits.CreditAmount < config.FreeTextSnapPrice {
			return false
		}
	}
	return true
}

func doDecreaseSnapCredits(c *fiber.Ctx, snapType string) error {
	user := c.Locals("user").(gofiberfirebaseauth.User)

	var userCredits model.UserCredits
	response := database.Pool.Where("user_id = ?", user.UserID).First(&userCredits)
	if err := database.ProcessDatabaseResponse(response); err != nil {
		return err
	}

	switch snapType {
	case "equation":
		response := database.Pool.Model(&model.UserCredits{}).Where("user_id = ?", user.UserID).Update("credit_amount", userCredits.CreditAmount-config.EquationTextSnapPrice)
		if err := database.ProcessDatabaseResponse(response); err != nil {
			return err
		}

		return addDecreaseSnapCreditsHistory(c, snapType, config.EquationTextSnapPrice)
	case "text":
		response := database.Pool.Model(&model.UserCredits{}).Where("user_id = ?", user.UserID).Update("credit_amount", userCredits.CreditAmount-config.FreeTextSnapPrice)
		if err := database.ProcessDatabaseResponse(response); err != nil {
			return err
		}

		return addDecreaseSnapCreditsHistory(c, snapType, config.FreeTextSnapPrice)
	}

	return nil
}

func addDecreaseSnapCreditsHistory(c *fiber.Ctx, snapType string, ammount float64) error {
	user := c.Locals("user").(gofiberfirebaseauth.User)

	var creditHistory model.CreditUsageHistory = model.CreditUsageHistory{
		UserID:      user.UserID,
		RequestType: snapType,
		Amount:      float64(ammount),
		Timestamp:   time.Now(),
	}

	response := database.Pool.Create(&creditHistory)
	return database.ProcessDatabaseResponse(response)
}
