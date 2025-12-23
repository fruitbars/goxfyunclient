package llmocr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"goxfyunclient/internal/service/llmocr/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockXFYunServer 创建一个模拟的讯飞 API 服务器
func mockXFYunServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestClient_RecognizeFile_Success(t *testing.T) {
	// 1. 设置模拟服务器
	server := mockXFYunServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 验证请求的基本信息
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// 构造成功的响应体
		// 模拟返回的 text 是 base64 编码的 json 字符串
		resultText := `{"result": "ok"}`
		encodedText := base64.StdEncoding.EncodeToString([]byte(resultText))

		response := models.ResponseBody{
			Header: models.ResponseHeader{
				Code:    0,
				Message: "Success",
				SID:     "test-sid-success",
			},
			Payload: models.ResponsePayload{
				Result: models.OCRResult{
					Text: encodedText,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// 2. 创建客户端，并将其 Host 指向模拟服务器
	client := NewClient("test-app-id", "test-api-key", "test-api-secret", WithHost(server.URL))

	// 3. 执行测试
	// 此处 imagePath 和 uid 仅为占位符，因为服务器返回的是固定响应
	result, err := client.RecognizeFile(context.Background(), "dummy-path.jpg", "test-uid")

	// 4. 断言结果
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	expectedText := `{"result": "ok"}`
	if result != expectedText {
		t.Errorf("Expected result text '%s', but got '%s'", expectedText, result)
	}
}

func TestClient_RecognizeFile_ApiError(t *testing.T) {
	// 1. 设置模拟服务器返回业务错误
	server := mockXFYunServer(t, func(w http.ResponseWriter, r *http.Request) {
		response := models.ResponseBody{
			Header: models.ResponseHeader{
				Code:    10106, // 模拟一个认证错误
				Message: "Invalid authorization",
				SID:     "test-sid-error",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 注意：即使业务失败，HTTP 状态码也可能是 200
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// 2. 创建客户端并指向模拟服务器
	client := NewClient("test-app-id", "test-api-key", "test-api-secret", WithHost(server.URL))

	// 3. 执行测试
	_, err := client.RecognizeFile(context.Background(), "dummy-path.jpg", "test-uid")

	// 4. 断言错误
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	expectedErrorMsg := "API返回错误: code=10106, message=Invalid authorization"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestClient_RecognizeFile_HttpError(t *testing.T) {
	// 1. 设置模拟服务器返回 HTTP 错误
	server := mockXFYunServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	defer server.Close()

	// 2. 创建客户端
	client := NewClient("test-app-id", "test-api-key", "test-api-secret", WithHost(server.URL))

	// 3. 执行测试
	_, err := client.RecognizeFile(context.Background(), "dummy-path.jpg", "test-uid")

	// 4. 断言错误
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	expectedErrorMsg := "执行OCR请求失败: 请求失败, 状态码: 500, 响应: Internal Server Error\n"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrorMsg, err.Error())
	}
}
