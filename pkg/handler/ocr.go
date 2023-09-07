package handler

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"log"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/vndee/lensquery-backend/pkg/model"
)

var (
	MathpixKey = os.Getenv("OCR_KEY")
	MathpixApp = os.Getenv("OCR_APP")
	MathpixURL = os.Getenv("OCR_URL")
)

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

	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		log.Printf("Failed to detect texts: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to detect texts")
	}

	if len(annotations) == 0 {
		log.Printf("No text found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "NO_TEXT_FOUND"})
	} else {
		log.Printf("Found %d text(s)", len(annotations))
		var annotationDescription []string
		for _, annotation := range annotations {
			log.Printf("Text: %q", annotation.Description)
			annotationDescription = append(annotationDescription, annotation.Description)
		}

		return c.Status(fiber.StatusOK).JSON(annotationDescription)
	}
}

func GetDocumentTextContent(c *fiber.Ctx) error {
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

	annotation, err := client.DetectDocumentText(ctx, image, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	if annotation == nil {
		log.Println("No text found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "NO_TEXT_FOUND"})
	}

	results := make(map[string]interface{})
	results["data"] = annotation.Text

	pages := make([]map[string]interface{}, len(annotation.Pages))
	for pageIndex, page := range annotation.Pages {
		pageData := map[string]interface{}{
			"Confidence": page.Confidence,
			"Width":      page.Width,
			"Height":     page.Height,
		}

		blocks := make([]map[string]interface{}, len(page.Blocks))
		for blockIndex, block := range page.Blocks {
			blockData := map[string]interface{}{
				"Confidence": block.Confidence,
				"BlockType":  block.BlockType,
			}

			paragraphs := make([]map[string]interface{}, len(block.Paragraphs))
			for paragraphIndex, paragraph := range block.Paragraphs {
				paragraphData := map[string]interface{}{
					"Confidence": paragraph.Confidence,
				}

				words := make([]map[string]interface{}, len(paragraph.Words))
				for wordIndex, word := range paragraph.Words {
					symbols := make([]string, len(word.Symbols))
					for i, s := range word.Symbols {
						symbols[i] = s.Text
					}
					wordText := strings.Join(symbols, "")
					words[wordIndex] = map[string]interface{}{
						"Confidence": word.Confidence,
						"Symbols":    wordText,
					}
				}
				paragraphData["Words"] = words
				paragraphs[paragraphIndex] = paragraphData
			}
			blockData["Paragraphs"] = paragraphs
			blocks[blockIndex] = blockData
		}
		pageData["Blocks"] = blocks
		pages[pageIndex] = pageData
	}

	results["pages"] = pages
	return c.Status(fiber.StatusOK).JSON(results)
}

func GetEquationTextContent(c *fiber.Ctx) error {
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

	return c.Status(response.StatusCode).JSON(responseData)
}
