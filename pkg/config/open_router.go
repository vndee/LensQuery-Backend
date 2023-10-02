package config

import (
	"net/http"

	openai "github.com/sashabaranov/go-openai"
)

var OpenRouterClient *openai.Client

type CustomRequestTransport struct {
	Origin http.RoundTripper
}

func (t *CustomRequestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("HTTP-Referer", "https://lensquery.com")
	req.Header.Set("X-Title", "LensQuery")
	return t.Origin.RoundTrip(req)
}

func SetupOpenRouterClient() {
	config := openai.DefaultConfig(OpenRouterAPIKey)
	config.BaseURL = OpenRouterEndpoint
	config.HTTPClient = &http.Client{
		Transport: &CustomRequestTransport{
			Origin: http.DefaultTransport,
		},
	}

	OpenRouterClient = openai.NewClientWithConfig(config)
}
