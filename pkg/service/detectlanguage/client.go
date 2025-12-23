package detectlanguage

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fruitbars/goxfyunclient/pkg/auth"
	"github.com/fruitbars/goxfyunclient/pkg/service/detectlanguage/models"
)

const (
	RequestURL = "https://cn-huadong-1.xf-yun.com/v1/private/s0ed5898e"
)

type Client struct {
	AppID      string
	APIKey     string
	APISecret  string
	Logger     *slog.Logger
	Host       string
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

// WithHost sets the host URL for the client.
func WithHost(host string) Option {
	return func(c *Client) {
		if host != "" {
			c.Host = host
		}
	}
}

func NewClient(appID, apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		AppID:     appID,
		APIKey:    apiKey,
		APISecret: apiSecret,
		Host:      RequestURL,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		HTTPClient: &http.Client{ // <--- 在这里初始化
			Timeout: 10 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Detect(text string) (string, error) {
	requestData := c.getRequestData(text)
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("请求数据JSON编码失败: %w", err)
	}

	authURL, err := auth.BuildAuthURL(c.Host, "POST", c.APIKey, c.APISecret, auth.SchemeTypeAPIKey)
	if err != nil {
		return "", fmt.Errorf("构建认证URL失败: %w", err)
	}
	if err != nil {
		return "", fmt.Errorf("构建认证URL失败: %w", err)
	}

	c.Logger.Debug("sending detectlanguage request", "url", authURL)

	req, err := http.NewRequest("POST", authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	//client := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("sending detectlanguage request failed", "url", authURL, "error", err)
		return "", fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	return c.dealResponse(resp)
}

func (c *Client) buildAuthRequestURL(requestURL, method string) (string, error) {
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	date := time.Now().UTC().Format(http.TimeFormat)

	signatureOrigin := fmt.Sprintf("host: %s\ndate: %s\n%s %s HTTP/1.1",
		parsedURL.Host,
		date,
		strings.ToUpper(method),
		parsedURL.RequestURI(),
	)

	mac := hmac.New(sha256.New, []byte(c.APISecret))
	mac.Write([]byte(signatureOrigin))
	signatureSha := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	authorizationOrigin := fmt.Sprintf(
		`api_key="%s", algorithm="%s", headers="%s", signature="%s"`,
		c.APIKey, "hmac-sha256", "host date request-line", signatureSha,
	)

	authorization := base64.StdEncoding.EncodeToString([]byte(authorizationOrigin))

	queryParams := url.Values{}
	queryParams.Add("host", parsedURL.Host)
	queryParams.Add("date", date)
	queryParams.Add("authorization", authorization)

	return fmt.Sprintf("%s?%s", requestURL, queryParams.Encode()), nil
}

func (c *Client) prepareReqData(text string) (models.RequestData, error) {
	data := c.getRequestData(text)
	data.Header.AppID = c.AppID
	data.Payload.Request.Text = base64.StdEncoding.EncodeToString([]byte(text))
	return data, nil
}

func (c *Client) dealResponse(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	c.Logger.Debug("received response", "status_code", resp.StatusCode, "body", string(body))

	var responseData models.ASELanguageDetectResponse
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", fmt.Errorf("error parsing response JSON: %w. Raw response: %s", err, string(body))
	}

	if responseData.Header.Code != 0 {
		c.Logger.Error("detectlanguage api returned an error",
			"code", responseData.Header.Code,
			"message", responseData.Header.Message,
			"sid", responseData.Header.Sid,
		)
		return "", fmt.Errorf("xfyun API error. Code: %d, Message: %s, sid: %s",
			responseData.Header.Code, responseData.Header.Message, responseData.Header.Sid)
	}

	if responseData.Payload.Result.Text == "" {
		return "", fmt.Errorf("empty result from xfyun API. sid: %s", responseData.Header.Sid)
	}

	decodedData, err := base64.StdEncoding.DecodeString(responseData.Payload.Result.Text)
	if err != nil {
		return "", fmt.Errorf("error decoding base64 content: %w", err)
	}

	var ldresult models.ASELanguageDetectTranResult
	err = json.Unmarshal(decodedData, &ldresult)
	if err != nil {
		return "", fmt.Errorf("error Unmarshal content: %w", err)
	}

	if len(ldresult.TransResult) <= 0 {
		return "", fmt.Errorf("TransResult is empty")
	}

	return ldresult.TransResult[0].LanProbs, nil
}

func (c *Client) getRequestData(text string) models.RequestData {
	b64Text := base64.StdEncoding.EncodeToString([]byte(text))
	uid := strings.ReplaceAll(uuid.New().String(), "-", "")
	return models.RequestData{
		Header: models.RequestHeader{
			AppID:  c.AppID, // Will be set in main
			UID:    uid,
			Status: 3,
		},
		Parameter: models.RequestParameter{
			Cnen: models.CnenParameter{
				Outfmt: "json",
				Result: struct {
					Encoding string `json:"encoding"`
					Compress string `json:"compress"`
					Format   string `json:"format"`
				}{
					Encoding: "utf8",
					Compress: "raw",
					Format:   "json",
				},
			},
		},
		Payload: models.RequestPayload{
			Request: struct {
				Encoding string `json:"encoding"`
				Compress string `json:"compress"`
				Format   string `json:"format"`
				Status   int    `json:"status"`
				Text     string `json:"text"`
			}{
				Encoding: "utf8",
				Compress: "raw",
				Format:   "plain",
				Status:   3,
				Text:     b64Text, // This will hold the base64 data
			},
		},
	}
}
