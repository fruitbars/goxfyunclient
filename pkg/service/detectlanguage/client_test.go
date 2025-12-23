package detectlanguage

import (
	"encoding/base64"
	"encoding/json"
	"goxfyunclient/internal/service/detectlanguage/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestClient_Detect_Success(t *testing.T) {
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 模拟成功的响应
		encodedResult, _ := json.Marshal(map[string]interface{}{
			"trans_result": []map[string]string{{"lan_probs": `{"cn": 1}`}},
		})
		encodedText := base64.StdEncoding.EncodeToString(encodedResult)

		resp := models.ASELanguageDetectResponse{
			Header: models.ResponseHeader{Code: 0, Message: "Success"},
			Payload: models.ResponsePayload{
				Result: models.ResultPayload{Text: encodedText},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	// 注入 mock server 的 URL
	client := NewClient("app-id", "api-key", "api-secret", WithHost(server.URL))
	result, err := client.Detect("你好")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != `{"cn": 1}` {
		t.Errorf("Expected result '{\"cn\": 1}', got '%s'", result)
	}
}

func TestClient_Detect_ApiError(t *testing.T) {
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		resp := models.ASELanguageDetectResponse{
			Header: models.ResponseHeader{Code: 10110, Message: "invalid text", Sid: "test-sid"},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	client := NewClient("app-id", "api-key", "api-secret", WithHost(server.URL))

	_, err := client.Detect("some text")

	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	expectedError := "xfyun API error. Code: 10110, Message: invalid text, sid: test-sid"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
