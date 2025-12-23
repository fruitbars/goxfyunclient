package ist

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/service/ist/models"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fruitbars/goxfyunclient/pkg/utils"
)

const (
	defaultHost     = "https://raasr.xfyun.cn/v2/api"
	apiUpload       = "/upload"
	apiGetResult    = "/getResult"
	pollingInterval = 5 * time.Second
)

// Client holds the configuration for the iFlytek API client.
type Client struct {
	AppID        string
	SecretKey    string
	HTTPClient   *http.Client
	UploadClient *http.Client
	Logger       *slog.Logger
	Host         string
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

// WithHTTPClients sets the http clients for upload and other requests.
func WithHTTPClients(uploadClient, defaultClient *http.Client) Option {
	return func(c *Client) {
		if uploadClient != nil {
			c.UploadClient = uploadClient
		}
		if defaultClient != nil {
			c.HTTPClient = defaultClient
		}
	}
}

// NewClient creates a new iFlytek LFAASR API client.
func NewClient(appID, secretKey string, opts ...Option) *Client {
	c := &Client{
		AppID:     appID,
		SecretKey: secretKey,
		Host:      defaultHost,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, // Set a reasonable timeout for general API calls
		},
		UploadClient: &http.Client{
			Timeout: 10 * time.Minute, // Set a longer timeout for file uploads
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

// getSignature generates the required signature for API authentication.
func (c *Client) getSignature(ts string) string {
	md5Hash := md5.New()
	md5Hash.Write([]byte(c.AppID + ts))
	md5Str := fmt.Sprintf("%x", md5Hash.Sum(nil))

	mac := hmac.New(sha1.New, []byte(c.SecretKey))
	mac.Write([]byte(md5Str))
	hmacSha1 := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(hmacSha1)
}

// UploadFile uploads the audio file to the LFAASR service.
func (c *Client) UploadFile(ctx context.Context, filePath string, options ...UploadOption) (string, error) {
	// The provided `filePath` is a URI, we need to handle different schemes.
	// For now, we will only handle local file URIs (file://) for simplicity.
	// And we'll treat it as a direct file path for now.
	// In a real implementation, you would parse the URI and download from http/smb.
	opts := &UploadOptions{}
	for _, opt := range options {
		opt(opts)
	}

	c.Logger.Debug("upload options", "audio_mode", opts.AudioMode, "audio_url", opts.AudioURL)

	var body io.Reader
	var fileSize, fileName string
	if opts.AudioMode == "urlLink" {
		if opts.AudioURL == "" {
			return "", fmt.Errorf("audioUrl is required when audioMode is urlLink")
		}
		if fileName == "" {
			// Derive fileName from audioUrl if not provided
			u, err := url.Parse(opts.AudioURL)
			if err != nil {
				return "", fmt.Errorf("invalid audioUrl: %w", err)
			}
			fileName = filepath.Base(u.Path)
		}
		body = nil
	} else {
		c.Logger.Debug("opening local file for upload", "path", filePath)
		file, err := os.Open(filePath)
		if err != nil {
			c.Logger.Error("failed to open file", "path", filePath, "error", err)
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			c.Logger.Error("failed to get file info", "path", filePath, "error", err)
			return "", fmt.Errorf("failed to get file info: %w", err)
		}

		c.Logger.Debug("file info", "name", fileInfo.Name(), "size", fileInfo.Size())

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			c.Logger.Error("failed to read file content", "path", filePath, "error", err)
			return "", fmt.Errorf("failed to read file content: %w", err)
		}
		body = bytes.NewReader(fileBytes)

		fileSize = strconv.FormatInt(fileInfo.Size(), 10)
		fileName = filepath.Base(filePath)
	}

	c.Logger.Debug("upload parameters", "file_size", fileSize, "file_name", fileName)

	ts := strconv.FormatInt(time.Now().Unix(), 10)

	params := url.Values{}
	params.Set("appId", c.AppID)
	params.Set("signa", c.getSignature(ts))
	params.Set("ts", ts)
	if fileSize != "" {
		params.Set("fileSize", fileSize)
	}
	if fileName != "" {
		params.Set("fileName", fileName)
	}
	if opts.Duration != "" {
		params.Set("duration", opts.Duration)
	} else {
		params.Set("duration", "200") // API requires duration, 0 means auto-detect
	}

	// Merge options parameters
	for key, values := range opts.ToURLValues() {
		if len(values) > 0 {
			params.Set(key, values[0]) // Use the first value from the slice
		}
	}

	uploadURL := c.Host + apiUpload + "?" + params.Encode()
	c.Logger.Info("uploading to url", "url", uploadURL)

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.UploadClient.Do(req)
	if err != nil {
		c.Logger.Error("failed to execute upload request", "url", uploadURL, "error", err)
		return "", fmt.Errorf("failed to execute upload request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("failed to read upload response body", "error", err)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	respBodyShort := utils.SafeSnippet(bodyBytes, 512)

	if resp.StatusCode != http.StatusOK {
		c.Logger.Error("upload request failed with non-200 status",
			"status", resp.Status,
			"response_body", respBodyShort,
		)
		return "", fmt.Errorf("Status: %s, response body: %s", resp.Status, respBodyShort)
	}

	c.Logger.Debug("upload response", "headers", resp.Header, "status", resp.Status, "body_snippet", respBodyShort)

	var uploadResp models.UploadResponse
	if err := json.Unmarshal(bodyBytes, &uploadResp); err != nil {
		c.Logger.Error("failed to decode upload json response", "body", respBodyShort, "error", err)
		return "", fmt.Errorf("failed to decode upload JSON response: %w", err)
	}

	if uploadResp.Code != "000000" {
		c.Logger.Error("upload api returned an error",
			"code", uploadResp.Code,
			"description", uploadResp.DescInfo,
		)
		return "", fmt.Errorf("upload failed with code %s: %s", uploadResp.Code, uploadResp.DescInfo)
	}

	return uploadResp.Content.OrderID, nil
}

// Process handles the entire synchronous transcription process: upload and get result.
func (c *Client) Process(ctx context.Context, filePath string, options ...UploadOption) (*models.GetResultResponse, error) {
	// Apply options to get resultType
	opts := &UploadOptions{}
	for _, opt := range options {
		opt(opts)
	}

	orderID, err := c.UploadFile(ctx, filePath, options...)
	if err != nil {
		c.Logger.Error("error during file upload step", "error", err)
		return nil, fmt.Errorf("error during file upload: %w", err)
	}
	c.Logger.Info("file uploaded successfully", "order_id", orderID)

	resultType := opts.GetResultType()
	result, err := c.GetTranscriptionResult(ctx, orderID, resultType)
	if err != nil {
		c.Logger.Error("error getting transcription result", "order_id", orderID, "result_type", resultType, "error", err)
		return nil, fmt.Errorf("error getting transcription result: %w", err)
	}

	return result, nil
}

// GetTranscriptionResult polls the API to get the final transcription result.
func (c *Client) GetTranscriptionResult(ctx context.Context, orderID string, resultType string) (*models.GetResultResponse, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sign := c.getSignature(ts)

	params := url.Values{}
	params.Set("appId", c.AppID)
	params.Set("signa", sign)
	params.Set("ts", ts)
	params.Set("orderId", orderID)
	params.Set("resultType", resultType)

	resultURL := c.Host + apiGetResult + "?" + params.Encode()

	// 立即执行一次，然后再开始轮询
	result, done, err := c.pollForResult(ctx, resultURL, orderID)
	if err != nil {
		c.Logger.Warn("initial poll for result failed, will start ticker", "order_id", orderID, "error", err)
	}
	if done {
		return result, err // 可能是成功，也可能是最终失败状态
	}

	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	consecutiveFailures := 0
	const maxConsecutiveFailures = 100 // 设置一个连续失败的上限

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-ticker.C:
			result, done, err := c.pollForResult(ctx, resultURL, orderID)
			if err != nil {
				consecutiveFailures++
				c.Logger.Warn("error polling for result, will retry", "order_id", orderID, "failures", consecutiveFailures, "error", err)
				if consecutiveFailures >= maxConsecutiveFailures {
					return nil, fmt.Errorf("查询结果失败次数过多，已达上限: %w", err)
				}
				continue // 继续下一次尝试
			}

			consecutiveFailures = 0 // 成功通信后重置计数器
			if done {
				return result, err // 返回最终结果（成功或失败）
			}
			// 如果没完成 (done == false)，则继续等待下一个 tick
		}
	}
}

// in client.go

// pollForResult 执行单次的结果轮询。
// 它返回最终结果、一个布尔值表示任务是否已终结（成功或失败），以及本次轮询遇到的任何错误。
func (c *Client) pollForResult(ctx context.Context, resultURL, orderID string) (result *models.GetResultResponse, done bool, err error) {
	// 1. 创建并发送 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", resultURL, nil)
	if err != nil {
		// 这是一个不可恢复的错误（对于本次尝试而言），应向上层报告
		return nil, false, fmt.Errorf("failed to create result request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.Logger.Debug("polling for transcription result", "url", resultURL, "order_id", orderID)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		// 网络错误，任务未终结，但本次轮询失败
		return nil, false, fmt.Errorf("failed to execute polling request: %w", err)
	}
	defer resp.Body.Close()

	// 2. 读取和解析响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// 读取响应体失败，任务未终结，但本次轮询失败
		return nil, false, fmt.Errorf("failed to read polling response body: %w", err)
	}

	bodySnippet := utils.SafeSnippet(body, 512)
	c.Logger.Debug("received polling response", "order_id", orderID, "body_snippet", bodySnippet)

	var resultResp models.GetResultResponse
	if err := json.Unmarshal(body, &resultResp); err != nil {
		// 解析 JSON 失败，任务未终结，但本次轮询失败
		return nil, false, fmt.Errorf("failed to decode result json: %w (body: %s)", err, bodySnippet)
	}

	// 3. 根据 API 返回的状态码进行逻辑判断
	status := resultResp.Content.OrderInfo.Status
	switch status {
	case 4: // 任务成功完成
		c.Logger.Info("transcription successful", "order_id", orderID)
		return &resultResp, true, nil // 返回结果, 标记为完成, 无错误

	case 3: // 任务仍在进行中
		c.Logger.Debug("transcription still in progress", "order_id", orderID)
		return nil, false, nil // 无结果, 标记为未完成, 无错误

	default: // 所有其他状态码均视为最终失败
		c.Logger.Error("transcription failed", "order_id", orderID, "status", status, "message", resultResp.DescInfo)
		finalErr := fmt.Errorf("transcription failed with status %d, message: %s", status, resultResp.DescInfo)
		return nil, true, finalErr // 无结果, 标记为完成, 有错误
	}
}
