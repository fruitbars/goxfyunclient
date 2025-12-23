package tts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/auth"
	"github.com/fruitbars/goxfyunclient/pkg/service/tts/models"
	"github.com/fruitbars/goxfyunclient/pkg/utils"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	APIHost     = "tts-api.xfyun.cn"
	APIEndpoint = "/v2/tts"
	Scheme      = "wss"
)

// Client TTSClient holds the configuration for the Text-to-Speech client.
type Client struct {
	AppID      string
	APIKey     string
	APISecret  string
	HTTPClient *http.Client
	conn       *websocket.Conn
	Logger     *slog.Logger
	testURL    string // for testing

	// 默认参数，可以在调用方法时被覆盖
	DefaultVoiceName   string
	DefaultAudioFormat string
}

// Option is a function that configures a TTSClient.
type Option func(*Client)

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.Logger = logger
		}
	}
}

// WithTestURL sets a test URL for the client, bypassing the standard URL builder.
// This is intended for testing purposes only.
func WithTestURL(url string) Option {
	return func(c *Client) {
		c.testURL = url
	}
}

func NewTTSClient(appID, apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		AppID:     appID,
		APIKey:    apiKey,
		APISecret: apiSecret,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		// 设置默认值
		DefaultVoiceName:   "x4_yezi",
		DefaultAudioFormat: "raw",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Connect establishes a WebSocket connection to the TTS service.
func (c *Client) Connect() error {
	if c.AppID == "" || c.APIKey == "" || c.APISecret == "" || strings.TrimSpace(c.AppID) == "" {
		return fmt.Errorf("AppID, APIKey, or APISecret is not configured")
	}

	//authURL, err := c.buildAuthURL()
	authURL, err := auth.AssembleAuthURLWithHostPath(Scheme, APIHost, APIEndpoint, "GET", c.APIKey, c.APISecret)
	c.Logger.Debug("connecting to tts websocket", "url", authURL)
	if err != nil {
		c.Logger.Error("could not build auth url", "error", err)
		return fmt.Errorf("could not build auth url: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, authURL, nil)
	if err != nil {
		c.Logger.Error("websocket dial failed", "url", authURL, "error", err)

		if resp != nil {
			bodyBytes, readBodyErr := io.ReadAll(resp.Body)
			if readBodyErr != nil {
				c.Logger.Error("failed to read response body", "error", readBodyErr)
				return err // 返回原始的拨号错误
			}
			respBodyShort := utils.SafeSnippet(bodyBytes, 512)
			c.Logger.Debug("WebSocket handshake response", "status", resp.Status, "headers", resp.Header, "body", respBodyShort)
			return fmt.Errorf("websocket dial failed: %w, status=%s, body=%s", err, resp.Status, respBodyShort)
		}

		return fmt.Errorf("websocket dial failed: %w", err)
	}
	c.Logger.Info("tts websocket connection established")

	c.conn = conn

	return nil
}

// Close closes the WebSocket connection.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// SendText sends a chunk of text to be synthesized.
// 更改 SendText 的函数签名
func (c *Client) SendText(text, voiceName, audioFormat string, isLast bool) error {
	if c.conn == nil {
		return fmt.Errorf("websocket connection is not established")
	}

	status := 1 // 中间帧
	if isLast {
		status = 2 // 最后一帧
	}

	// 动态构建业务参数
	business := models.RequestBusiness{VCN: voiceName, TTE: "UTF8"}
	switch audioFormat {
	case "mp3":
		business.AUE = "lame"
		sfl := 1
		business.SFL = &sfl
	case "raw":
		fallthrough
	default:
		business.AUE = "raw"
		business.AUF = "audio/L16;rate=16000"
	}

	payload := models.RequestPayload{
		Common:   models.RequestCommon{AppID: c.AppID},
		Business: business,
		Data: models.RequestData{
			Status: status,
			Text:   base64.StdEncoding.EncodeToString([]byte(text)),
		},
	}
	return c.conn.WriteJSON(payload)
}

// ReceiveAudio receives audio data from the WebSocket connection.
// It returns the audio data, a flag indicating if it's the last frame, and any error.
func (c *Client) ReceiveAudio() ([]byte, bool, error) {
	if c.conn == nil {
		return nil, false, fmt.Errorf("websocket connection is not established")
	}

	_, message, err := c.conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			return nil, true, nil // Clean close
		}
		return nil, false, fmt.Errorf("error reading message: %w", err)
	}

	var resp models.ServerResponse
	if err := json.Unmarshal(message, &resp); err != nil {
		c.Logger.Error("failed to unmarshal tts response", "message", string(message), "error", err)
		return nil, false, err
	}

	if resp.Code != 0 {
		c.Logger.Error("tts api returned an error",
			"code", resp.Code,
			"message", resp.Message,
			"sid", resp.SID,
		)
		return nil, false, fmt.Errorf("server error: code=%d, message=%s, sid=%s", resp.Code, resp.Message, resp.SID)
	}

	audioData, err := base64.StdEncoding.DecodeString(resp.Data.Audio)
	if err != nil {
		c.Logger.Error("failed to decode base64 audio data", "error", err)
		return nil, false, err
	}

	isLast := resp.Data.Status == 2
	return audioData, isLast, nil
}

// client.go

// TextToSpeech 执行一次性的文本到语音转换。
// 它是 StreamTextReader 的一个便捷封装。
func (c *Client) TextToSpeech(ctx context.Context, text, voiceName, audioFormat string) ([]byte, error) {
	// 使用已有的流式方法来获取一个 reader
	reader, err := c.StreamTextReader(ctx, text, voiceName, audioFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to start TTS stream: %w", err)
	}
	defer reader.Close()

	// 将流中的所有数据读取到一个字节切片中
	audioData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio stream: %w", err)
	}

	return audioData, nil
}

func (c *Client) StreamTextReader(ctx context.Context, text, voiceName, audioFormat string) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	if err := c.Connect(); err != nil {
		return nil, err
	}

	go func() {
		defer func() {
			c.Close()
			_ = pw.Close()
		}()
		// 发送首帧/或分段 SendText
		if err := c.SendText(text, voiceName, audioFormat, true); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		for {
			select {
			case <-ctx.Done():
				_ = pw.CloseWithError(ctx.Err())
				return
			default:
			}
			audio, last, err := c.ReceiveAudio()
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
			if len(audio) > 0 {
				if _, werr := pw.Write(audio); werr != nil {
					_ = pw.CloseWithError(werr)
					return
				}
			}
			if last {
				return
			}
		}
	}()

	return pr, nil
}
