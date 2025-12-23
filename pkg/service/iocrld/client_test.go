package iocrld

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"goxfyunclient/internal/service/iocrld/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Process_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟成功响应
		responseText := `{"pages": [{"angle": 0}]}`
		encodedText := base64.StdEncoding.EncodeToString([]byte(responseText))

		resp := models.Response{
			Header: models.Header{Code: 0, Message: "Success", SID: "test-sid"},
			Payload: models.Payload{
				JSON:  models.JSONPayload{Text: encodedText},
				Image: models.ImagePayload{Image: "dummy-image-base64"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("app-id", "api-key", "api-secret", WithHost(server.URL))
	result, err := client.Process(context.Background(), "track-id", "dummy-pic-base64", nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.SID != "test-sid" {
		t.Errorf("Expected sid 'test-sid', got '%s'", result.SID)
	}
	if result.Text != `{"pages": [{"angle": 0}]}` {
		t.Errorf("Unexpected result text: %s", result.Text)
	}
	if result.ImageBase64 != "dummy-image-base64" {
		t.Error("ImageBase64 was not correctly extracted from response")
	}
}

func TestClient_Process_ApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.Response{
			Header: models.Header{Code: 10110, Message: "some error", SID: "error-sid"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("app-id", "api-key", "api-secret", WithHost(server.URL))
	_, err := client.Process(context.Background(), "track-id", "dummy-pic-base64", nil)

	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected error of type APIError, but got %T", err)
	}
	if apiErr.Code != 10110 {
		t.Errorf("Expected error code 10110, got %d", apiErr.Code)
	}
	if apiErr.Message != "some error" {
		t.Errorf("Expected error message 'some error', got '%s'", apiErr.Message)
	}
}
