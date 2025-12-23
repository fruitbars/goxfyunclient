package translate

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/auth"
	"github.com/fruitbars/goxfyunclient/pkg/service/translate/models"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	defaultHost = "https://itrans.xf-yun.com/v1/its"
)

// Client for the Xunfei translation service.
type Client struct {
	HostURL    string
	AppID      string
	APIKey     string
	APISecret  string
	Logger     *slog.Logger
	HTTPClient *http.Client
}

// Option is a function that configures a Client.
type Option func(*Client)

// WithHost sets the host for the client.
func WithHost(host string) Option {
	return func(c *Client) {
		if host != "" {
			c.HostURL = host
		}
	}
}

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.Logger = logger
		}
	}
}

func NewClient(appID, apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		HostURL:   defaultHost,
		AppID:     appID,
		APIKey:    apiKey,
		APISecret: apiSecret,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Translate performs the translation using RESTful API.
func (c *Client) Translate(ctx context.Context, text, from, to string) (string, error) {
	authURL, err := auth.AssembleAuthURL(c.HostURL, "POST", c.APIKey, c.APISecret)
	if err != nil {
		return "", fmt.Errorf("构建认证URL失败: %w", err)
	}

	requestBody, err := c.buildRequestBody(text, from, to)
	if err != nil {
		return "", fmt.Errorf("构建请求体失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.Logger.Debug("sending translate request", "url", authURL, "from", from, "to", to)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("sending translate request failed", "url", authURL, "error", err)
		return "", fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("reading translate response body failed", "error", err)
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}

	return c.parseResponse(responseBody)
}

func (c *Client) buildRequestBody(text, from, to string) ([]byte, error) {
	encodedText := base64.StdEncoding.EncodeToString([]byte(text))

	reqBody := models.RequestBody{
		Header: models.RequestHeader{
			AppID:  c.AppID,
			Status: 3,
		},
		Parameter: models.RequestParameter{
			ITS: models.RequestParameterITS{
				From:   from,
				To:     to,
				Result: make(map[string]interface{}), // Per doc, this is an empty object
			},
		},
		Payload: models.RequestPayload{
			InputData: models.RequestInputData{
				Encoding: "utf8",
				Status:   3,
				Text:     encodedText,
			},
		},
	}
	return json.Marshal(reqBody)
}

func (c *Client) parseResponse(body []byte) (string, error) {
	var respData models.ResponseBody
	if err := json.Unmarshal(body, &respData); err != nil {
		c.Logger.Error("unmarshalling translate response failed", "body", string(body), "error", err)
		return "", fmt.Errorf("解析JSON响应失败: %w", err)
	}

	if respData.Header.Code != 0 {
		c.Logger.Error("translate api returned an error",
			"code", respData.Header.Code,
			"message", respData.Header.Message,
			"sid", respData.Header.Sid,
		)
		return "", fmt.Errorf("API返回错误: code=%d, message=%s, sid=%s",
			respData.Header.Code, respData.Header.Message, respData.Header.Sid)
	}

	decodedText, err := base64.StdEncoding.DecodeString(respData.Payload.Result.Text)
	if err != nil {
		c.Logger.Error("decoding result text failed", "error", err)
		return "", fmt.Errorf("base64解码结果失败: %w", err)
	}

	// The decoded text is another JSON object. We need to extract the `dst` field.
	var resultData models.EngineResultPayload
	if err := json.Unmarshal(decodedText, &resultData); err != nil {
		c.Logger.Error("unmarshalling decoded text failed", "decoded_text", string(decodedText), "error", err)
		// Fallback to returning the raw decoded string if it's not the expected JSON format
		return string(decodedText), nil
	}

	return resultData.TransResult.Dst, nil
}
