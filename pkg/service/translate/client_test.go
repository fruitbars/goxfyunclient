package translate

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Translate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Verify request body
		var reqBody RequestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if reqBody.Parameter.ITS.From != "en" || reqBody.Parameter.ITS.To != "cn" {
			t.Errorf("Unexpected from/to languages: %s/%s", reqBody.Parameter.ITS.From, reqBody.Parameter.ITS.To)
		}

		// 2. Mock success response
		resultText := `{"trans_result": {"dst": "你好"}}`
		encodedResult := base64.StdEncoding.EncodeToString([]byte(resultText))

		resp := ResponseBody{
			Header: ResponseHeader{Code: 0, Message: "Success", Sid: "sid-success"},
			Payload: ResponsePayload{
				Result: ResponseResult{Text: encodedResult},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("app-id", "api-key", "api-secret", WithHost(server.URL))
	result, err := client.Translate(context.Background(), "hello", "en", "cn")

	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if result != "你好" {
		t.Errorf("Expected '你好', got '%s'", result)
	}
}

func TestClient_Translate_ApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ResponseBody{
			Header: ResponseHeader{Code: 10106, Message: "invalid auth", Sid: "sid-error"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("app-id", "api-key", "api-secret", WithHost(server.URL))
	_, err := client.Translate(context.Background(), "hello", "en", "cn")

	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	expectedError := "API返回错误: code=10106, message=invalid auth, sid=sid-error"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
