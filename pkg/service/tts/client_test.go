package tts

import (
	"context"
	"encoding/base64"
	"goxfyunclient/internal/service/tts/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func mockWebSocketServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()
		handler(conn)
	}))
	return server
}

func TestTTSClient_TextToSpeech_Success(t *testing.T) {
	server := mockWebSocketServer(t, func(conn *websocket.Conn) {
		var req models.RequestPayload
		if err := conn.ReadJSON(&req); err != nil {
			t.Errorf("Failed to read JSON request: %v", err)
			return
		}

		expectedText := base64.StdEncoding.EncodeToString([]byte("hello"))
		if req.Data.Text != expectedText {
			t.Errorf("Expected text '%s', got '%s'", expectedText, req.Data.Text)
		}

		// Send back a response
		audioBase64 := base64.StdEncoding.EncodeToString([]byte("dummy-audio-data"))
		resp := models.ServerResponse{
			Code:    0,
			Message: "success",
			SID:     "tts-sid-success",
			Data: struct {
				Audio  string `json:"audio"`
				Status int    `json:"status"`
				CED    string `json:"ced"`
			}{Status: 2, Audio: audioBase64},
		}
		if err := conn.WriteJSON(resp); err != nil {
			t.Errorf("Failed to write JSON response: %v", err)
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewTTSClient("app-id", "api-key", "api-secret", WithTestURL(wsURL))

	audioData, err := client.TextToSpeech(context.Background(), "hello", "xiaoyan", "mp3")
	if err != nil {
		t.Fatalf("TextToSpeech failed: %v", err)
	}
	if string(audioData) != "dummy-audio-data" {
		t.Errorf("Expected 'dummy-audio-data', got '%s'", string(audioData))
	}
}

func TestTTSClient_TextToSpeech_ApiError(t *testing.T) {
	server := mockWebSocketServer(t, func(conn *websocket.Conn) {
		resp := models.ServerResponse{
			Code:    10106, // Invalid parameter error code
			Message: "invalid parameter",
			SID:     "tts-sid-error",
		}
		conn.WriteJSON(resp)
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewTTSClient("app-id", "api-key", "api-secret", WithTestURL(wsURL))

	_, err := client.TextToSpeech(context.Background(), "hello", "invalid-voice", "mp3")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	expectedError := "API返回错误: code=10106, message=invalid parameter"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}
