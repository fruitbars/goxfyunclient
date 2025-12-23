package ocr

import (
	"context"
	"encoding/json"
	"goxfyunclient/internal/service/ocr/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Recognize_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟成功响应
		resp := models.Response{
			Header: models.Header{Code: 0, Message: "Success", Sid: "test-sid-success"},
			Payload: models.Payload{
				Result: models.Result{
					Text: "eyJzdGF0dXMiOiAzLCAicGFnZXMiOiBbeyJhbmdsZSI6IDAsICJoZWlnaHQiOiAxMjAwLCAid2lkdGgiOiA5MDAsICJjb250ZW50IjogW3siY29udGVudCI6ICJUZXN0IiwgInBvc2l0aW9uIjogWzAsIDAsIDkwMCwgMF0gfV0gfV19", // base64 of `{"status": 3, "pages": [{"angle": 0, "height": 1200, "width": 900, "content": [{"content": "Test", "position": [0, 0, 900, 0]}]}]}`
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("app-id", "api-key", "api-secret", WithHost(server.URL))

	result, err := client.Recognize(context.Background(), "dummy-base64-image", General)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Header.Sid != "test-sid-success" {
		t.Errorf("Expected sid 'test-sid-success', got '%s'", result.Header.Sid)
	}
	if result.Payload.Result.Text == "" {
		t.Error("Expected result text to be non-empty")
	}
}

func TestClient_Recognize_ApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟 API 错误响应
		resp := models.Response{
			Header: models.Header{Code: 10106, Message: "Invalid parameter", Sid: "test-sid-error"},
		}
		w.Header().Set("Content-Type", "application/json")
		// 讯飞接口即使业务失败，HTTP状态码也可能是200
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("app-id", "api-key", "api-secret", WithHost(server.URL))

	_, err := client.Recognize(context.Background(), "dummy-base64-image", General)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
	expectedError := "API返回错误: code=10106, message=Invalid parameter, sid=test-sid-error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}
