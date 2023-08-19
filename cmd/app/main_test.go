package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexRoute(t *testing.T) {
	tests := []struct {
		description   string
		method        string
		path          string
		expectedCode  int
		expectedError string
		expectedBody  string
	}{
		{
			description:   "GET /",
			method:        "GET",
			path:          "/",
			expectedCode:  http.StatusBadRequest,
			expectedError: "",
			expectedBody:  "Missing or malformed Token",
		},
		{
			description:   "GET /healthcheck",
			method:        "GET",
			path:          "/",
			expectedCode:  http.StatusBadRequest,
			expectedError: "",
			expectedBody:  "Missing or malformed Token",
		},
		{
			description:   "GET /api/v1/ocr/get_access_token",
			method:        "GET",
			path:          "/api/v1/ocr/get_access_token",
			expectedCode:  http.StatusBadRequest,
			expectedError: "",
			expectedBody:  "Missing or malformed Token",
		},
	}

	app := Setup()

	for _, tc := range tests {
		req, _ := http.NewRequest(tc.method, tc.path, nil)
		resp, err := app.Test(req)

		assert.Nil(t, err, tc.description)
		assert.Equal(t, tc.expectedCode, resp.StatusCode, tc.description)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, tc.expectedBody, string(body), tc.description)
	}
}

func TestFirebaseAuthentication(t *testing.T) {

}

func TestGetAccessToken(t *testing.T) {

}
