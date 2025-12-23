package iocrld

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/auth"
	"github.com/fruitbars/goxfyunclient/pkg/service/iocrld/models"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	defaultHost = "https://cn-huabei-1.xf-yun.com/v1/private/s15fc3900"
)

// Client 封装了讯飞私有接口调用（签名、请求、解析）
type Client struct {
	AppID      string
	APIKey     string
	APISecret  string
	Host       string
	Logger     *slog.Logger
	HTTPClient *http.Client
}

// Option is a function that configures a Client.
type Option func(*Client)

// WithHost sets the host for the client.
func WithHost(host string) Option {
	return func(c *Client) {
		if host != "" {
			c.Host = host
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

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.HTTPClient = client
		}
	}
}

func NewClient(appID, apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		AppID:     appID,
		APIKey:    apiKey,
		APISecret: apiSecret,
		Host:      defaultHost,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Result 为对外结果载体
type Result struct {
	SID         string // 讯飞返回的 sid
	ImageBase64 string // 处理后的图片（base64）
	Text        string // 解析后的文本（已尽力从 base64 解码为字符串）
}

// APIError 代表讯飞返回的业务错误（header.code != 0）
type APIError struct {
	Code    int
	Message string
	SID     string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("iFlytek API error: code=%d, sid=%s, message=%s", e.Code, e.SID, e.Message)
}

func (c *Client) Process(ctx context.Context, trackID string, picBase64 string, customParams map[string]interface{}) (*models.Response, error) {
	if strings.TrimSpace(c.AppID) == "" || strings.TrimSpace(c.APIKey) == "" || strings.TrimSpace(c.APISecret) == "" {
		return nil, fmt.Errorf("missing credentials: AppID/APIKey/APISecret are required")
	}
	if strings.TrimSpace(c.Host) == "" {
		return nil, fmt.Errorf("missing endpoint")
	}
	if strings.TrimSpace(picBase64) == "" {
		return nil, fmt.Errorf("pictureBase64 is empty")
	}

	var jsonDataRaw json.RawMessage
	if customParams != nil {
		jb, err := json.Marshal(customParams)
		if err != nil {
			return nil, fmt.Errorf("invalid customParams: %w", err)
		}
		jsonDataRaw = jb
	} else {
		jsonDataRaw = json.RawMessage(`{}`)
	}

	// --- 1) 组装请求 ---
	reqBody := models.Request{
		Header: models.Header{
			AppID:     c.AppID,
			RequestID: &trackID,
			Status:    3,
		},
		Parameter: models.Parameter{
			IOCRld: models.IOCRld{
				JSON: models.JSONParams{
					Encoding: "utf8",
					Compress: "raw",
					Format:   "json",
				},
				Image: models.ImageParams{
					Encoding: "jpg",
				},
			},
		},
		Payload: models.Payload{
			JSON: models.JSONPayload{
				Encoding: "utf8",
				Compress: "raw",
				Format:   "json",
				Status:   3,
				Text:     base64JSON(jsonDataRaw),
			},
			Image: models.ImagePayload{
				Encoding: "jpg",
				Status:   3,
				Image:    picBase64,
			},
		},
	}

	authURL, err := auth.BuildAuthURL(c.Host, "POST", c.APIKey, c.APISecret, auth.SchemeTypeAPIKey)
	if err != nil {
		return nil, fmt.Errorf("build auth url failed: %w", err)
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	c.Logger.Debug("sending iocrld request", "url", authURL)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("sending iocrld request failed", "url", authURL, "error", err)
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.Logger.Debug("received iocrld response", "status_code", resp.StatusCode, "body_snippet", string(b[:min(512, len(b))]))

	// --- 3) 解析响应 ---
	var ifResp models.Response
	if err := json.Unmarshal(b, &ifResp); err != nil {
		return nil, fmt.Errorf("unmarshal iflytek response failed: %w", err)
	}
	if ifResp.Header.Code != 0 {
		c.Logger.Error("iocrld api returned an error",
			"code", ifResp.Header.Code,
			"message", ifResp.Header.Message,
			"sid", ifResp.Header.SID,
		)
		return nil, &APIError{
			Code:    ifResp.Header.Code,
			Message: ifResp.Header.Message,
			SID:     ifResp.Header.SID,
		}
	}

	return &ifResp, nil
}

func base64JSON(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
