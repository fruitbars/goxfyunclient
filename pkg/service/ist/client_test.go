package ist

import (
	"context"
	"encoding/json"
	"goxfyunclient/internal/service/ist/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_Process_Success(t *testing.T) {
	orderID := "test-order-id-success"
	pollingCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, apiUpload) {
			// 模拟上传成功
			resp := models.UploadResponse{
				Code: "000000",
				Content: struct {
					OrderID string `json:"orderId"`
				}{OrderID: orderID},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(r.URL.Path, apiGetResult) {
			pollingCount++
			var status int
			var result string
			// 模拟轮询：前两次返回处理中，第三次返回成功
			if pollingCount < 3 {
				status = 3 // Processing
			} else {
				status = 4 // Success
				result = `{"key": "value"}`
			}

			resp := models.GetResultResponse{
				Code: "000000",
				Content: models.GetResultContent{
					OrderInfo:   models.OrderInfo{Status: status},
					OrderResult: result,
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
	}))
	defer server.Close()

	// 替换全局 Host 变量
	originalPollingInterval := pollingInterval
	pollingInterval = 10 * time.Millisecond // 加速测试
	defer func() { pollingInterval = originalPollingInterval }()

	client := NewClient("app-id", "secret-key", WithHost(server.URL))
	result, err := client.Process(context.Background(), "dummy.mp3")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if pollingCount != 3 {
		t.Errorf("Expected to poll 3 times, but polled %d times", pollingCount)
	}
	if result.Content.OrderResult != `{"key": "value"}` {
		t.Errorf("Unexpected order result: %s", result.Content.OrderResult)
	}
}

func TestClient_Process_UploadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟上传失败
		resp := models.UploadResponse{Code: "10001", DescInfo: "upload failed"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("app-id", "secret-key", WithHost(server.URL))
	_, err := client.Process(context.Background(), "dummy.mp3")

	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	expectedError := "error during file upload: upload failed with code 10001: upload failed"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
