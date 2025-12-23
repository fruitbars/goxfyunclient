package llmocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/auth"
	"github.com/fruitbars/goxfyunclient/pkg/service/llmocr/models"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	HOST = "https://cbm01.cn-huabei-1.xf-yun.com/v1/private/se75ocrbm"
)

// Client represents the llmocr client
type Client struct {
	AppID      string
	ApiKey     string
	ApiSecret  string
	Host       string
	Logger     *slog.Logger
	HTTPClient *http.Client // <--- 添加此字段
}

// Option is a function that configures a Client.
type Option func(*Client)

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.Logger = logger
		}
	}
}

// WithHost sets the host for the client.
func WithHost(host string) Option {
	return func(c *Client) {
		if host != "" {
			c.Host = host
		}
	}
}

// NewClient creates a new llmocr client.
func NewClient(appID, apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		AppID:     appID,
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
		Host:      HOST,
		// Default to a silent logger
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		HTTPClient: &http.Client{ // <--- 在这里初始化
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// RecognizeFile 静态图片识别，从文件路径读取
func (c *Client) RecognizeFile(ctx context.Context, imagePath, uid string) (string, error) {
	// 1. 读取并编码图片
	base64Image, imageType, err := readAndEncodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("读取图片文件失败: %w", err)
	}
	return c.ocr(ctx, uid, base64Image, imageType)
}

// RecognizeBytes 静态图片识别，从字节流读取
func (c *Client) RecognizeBytes(ctx context.Context, imageData []byte, imageType, uid string) (string, error) {
	if imageType == "" {
		contentType := http.DetectContentType(imageData)
		switch {
		case strings.Contains(contentType, "jpeg"):
			imageType = "jpg"
		case strings.Contains(contentType, "png"):
			imageType = "png"
		// ... 其他类型
		default:
			imageType = "jpg"
		}
	}
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	return c.ocr(ctx, uid, base64Image, imageType)
}

// ocr 是执行OCR的核心私有方法
func (c *Client) ocr(ctx context.Context, uid, base64Image, imageType string) (string, error) {
	// 1. 构建请求体
	requestBody := c.buildRequestBody(uid, base64Image, imageType)
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 2. 执行请求
	responseBytes, err := c.executeOCRRequest(ctx, requestBytes)
	if err != nil {
		return "", fmt.Errorf("执行OCR请求失败: %w", err)
	}

	// 3. 解析响应并返回结果
	return c.parseResponse(responseBytes)
}

// readAndEncodeImage 读取图片文件，返回Base64编码字符串和文件类型
func readAndEncodeImage(path string) (string, string, error) {
	imgBytes, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	base64Str := base64.StdEncoding.EncodeToString(imgBytes)
	fileType := strings.TrimPrefix(filepath.Ext(path), ".")
	return base64Str, fileType, nil
}

// buildRequestBody 使用结构体构建请求体
func (c *Client) buildRequestBody(uid, imageBase64, fileType string) models.RequestBody {
	return models.RequestBody{
		Header: models.Header{
			AppID:  c.AppID,
			UID:    uid,
			Status: 0,
		},
		Parameter: models.Parameter{
			OCR: models.OCRParams{
				ResultOption: "normal",
				ResultFormat: "json,markdown,sed,word", // 请求json格式以获得结构化数据
				OutputType:   "one_shot",
				Result: models.ResultParam{
					Encoding: "utf8",
					Compress: "raw",
					Format:   "plain",
				},
			},
		},
		Payload: models.Payload{
			Image: models.ImagePayload{
				Encoding: fileType,
				Image:    imageBase64,
				Status:   0, // 标记为最后一块数据
			},
		},
	}
}

// executeOCRRequest 负责签名和发送HTTP请求
func (c *Client) executeOCRRequest(ctx context.Context, payload []byte) ([]byte, error) {
	// 1. 生成带鉴权的URL
	//authURL, err := c.assembleRequestUrl("POST")
	authURL, err := auth.BuildAuthURL(HOST, "POST", c.ApiKey, c.ApiSecret, auth.SchemeTypeHMAC)
	if err != nil {
		return nil, fmt.Errorf("生成鉴权URL失败: %w", err)
	}

	// 2. 创建带上下文的HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.Logger.Debug("sending llmocr request", "url", authURL, "uid", req.Header.Get("uid"))

	// 3. 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 4. 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 5. 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.Logger.Error("llmocr request failed",
			"status_code", resp.StatusCode,
			"response", string(responseBody),
		)
		return nil, fmt.Errorf("请求失败, 状态码: %d, 响应: %s", resp.StatusCode, string(responseBody))
	}

	c.Logger.Debug("llmocr request successful")
	return responseBody, nil
}

// parseResponse 解析API返回的JSON数据
func (c *Client) parseResponse(responseBytes []byte) (string, error) {
	var respData models.ResponseBody
	if err := json.Unmarshal(responseBytes, &respData); err != nil {
		return "", fmt.Errorf("解析响应JSON失败: %w", err)
	}

	// 检查API返回的业务错误码
	if respData.Header.Code != 0 {
		c.Logger.Error("llmocr api returned an error",
			"code", respData.Header.Code,
			"message", respData.Header.Message,
			"sid", respData.Header.SID,
		)
		return "", fmt.Errorf("API返回错误: code=%d, message=%s", respData.Header.Code, respData.Header.Message)
	}

	// Base64解码最终的文本结果
	decodedText, err := base64.StdEncoding.DecodeString(respData.Payload.Result.Text)
	if err != nil {
		return "", fmt.Errorf("Base64解码结果失败: %w", err)
	}

	return string(decodedText), nil
}
