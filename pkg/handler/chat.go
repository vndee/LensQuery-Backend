package handler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/sashabaranov/go-openai"
	"github.com/valyala/fasthttp"
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

func Completion(c *fiber.Ctx) error {
	var requestBody *openai.ChatCompletionRequest
	if err := c.BodyParser(&requestBody); err != nil {
		log.Println("BodyParser:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, "keep-alive")
	c.Set(fiber.HeaderAccessControlAllowOrigin, "*")
	c.Set(fiber.HeaderTransferEncoding, "chunked")

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		stream, err := config.OpenRouterClient.CreateChatCompletionStream(context.Background(), *requestBody)
		if err != nil {
			log.Println("CreateChatCompletionStream:", err)
			return
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()

			if errors.Is(err, io.EOF) {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				err = w.Flush()
				if err != nil {
					fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)
				}

				stream.Close()
				break
			}

			fmt.Fprintf(w, "data: %s\n\n", response.Choices[0].Delta.Content)

			err = w.Flush()
			if err != nil {
				// Refreshing page in web browser will establish a new
				// SSE connection, but only (the last) one is alive, so
				// dead connections must be closed here.
				fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)
				stream.Close()
				break
			}
		}
	}))

	return nil
}
