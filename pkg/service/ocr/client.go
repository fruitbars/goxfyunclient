package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/auth"
	"github.com/fruitbars/goxfyunclient/pkg/utils"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const apiURL = "https://cn-east-1.api.xf-yun.com/v1/ocr"

// 压缩策略
const (
	MaxUncompressedBytes  = int64(7.5 * 1024 * 1024) // 超过则压缩
	TargetCompressedBytes = 2 * 1024 * 1024          // 目标 2MB
)

// Client holds credentials and HTTP client.
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
		Host:      apiURL,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, // 给个总超时更安全
			// Transport: 自定义的话可在 Option 里扩展
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	// 防御：确保最终不是 nil
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
	return c
}

// RecognizedText decodes base64 "text" field.
func (r *OcrResponse) RecognizedText() (string, error) {
	if r.Payload.OcrOutputText.Text == "" {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(r.Payload.OcrOutputText.Text)
	if err != nil {
		return "", fmt.Errorf("failed to decode recognized text: %w", err)
	}
	return string(decoded), nil
}

func (c *Client) validateCredentials() error {
	if c.AppID == "" || c.APIKey == "" || c.APISecret == "" {
		return fmt.Errorf("AppID/APIKey/APISecret is not configured")
	}
	return nil
}

// 重构后的 buildAuthURL
func (c *Client) buildAuthURL() (string, error) {
	if c.APIKey == "" || c.APISecret == "" {
		return "", fmt.Errorf("missing APIKey/APISecret")
	}
	// 直接调用通用鉴权函数
	u, err := url.Parse(c.Host) // 使用 c.Host 而不是硬编码的 apiURL
	if err != nil {
		return "", fmt.Errorf("invalid host url: %w", err)
	}
	return auth.AssembleAuthURLWithHostPath(u.Scheme, u.Host, u.Path, "POST", c.APIKey, c.APISecret)
}

// ----------- Public APIs -----------

// RecognizePath reads file then calls RecognizeBytes.
func (c *Client) RecognizePath(ctx context.Context, imagePath, imgEncoding, language string) (*OcrResponse, error) {
	if err := c.validateCredentials(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("read image file '%s' failed: %w", imagePath, err)
	}
	return c.RecognizeBytes(ctx, data, imgEncoding, language)
}

// RecognizeBytes sends the image bytes to API.
func (c *Client) RecognizeBytes(ctx context.Context, image []byte, imgEncoding, language string) (*OcrResponse, error) {
	if err := c.validateCredentials(); err != nil {
		return nil, err
	}
	if len(image) == 0 {
		return nil, fmt.Errorf("empty image data")
	}
	if imgEncoding == "" {
		imgEncoding = "jpg"
	}
	if language == "" {
		// 按需修改：示例兼容中英
		language = "cn|en"
	}

	encoded := base64.StdEncoding.EncodeToString(image)

	// Build JSON body
	var body requestBody
	body.Header = reqHeader{
		AppID:  c.AppID,
		Status: 3,
	}
	body.Parameter.OCR.Language = language
	body.Parameter.OCR.OcrOutputText.Encoding = "utf8"
	body.Parameter.OCR.OcrOutputText.Compress = "raw"
	body.Parameter.OCR.OcrOutputText.Format = "json"
	body.Payload.Image.Encoding = imgEncoding
	body.Payload.Image.Image = encoded
	body.Payload.Image.Status = 3

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %w", err)
	}

	payloadShort := utils.SafeSnippet(payload, 512)
	c.Logger.Debug("building ocr request body", "payloadShort", payloadShort)

	authURL, err := c.buildAuthURL()
	if err != nil {
		return nil, fmt.Errorf("build auth url failed: %w", err)
	}

	c.Logger.Debug("sending ocr request", "url", authURL, "category", language)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewReader(payload))
	if err != nil {
		c.Logger.Error("create request failed", "error", err)
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	u, _ := url.Parse(c.Host)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", u.Host)    // keep same as demo
	req.Header.Set("App_id", c.AppID) // casing per demo

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("send request failed", "error", err)
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("read response failed", "error", err)
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	respBodyShort := utils.SafeSnippet(respBytes, 512)

	c.Logger.Debug("received ocr response", "status_code", resp, "body", respBodyShort)

	var ocrResp OcrResponse
	if err := json.Unmarshal(respBytes, &ocrResp); err != nil {
		c.Logger.Error("unmarshal response failed", "body", string(respBytes), "error", err)
		return nil, fmt.Errorf("unmarshal response failed: %w (body: %s)", err, string(respBytes))
	}
	if ocrResp.Header.Code != 0 {
		c.Logger.Error("API error",
			"code", ocrResp.Header.Code,
			"message", ocrResp.Header.Message,
		)
		return nil, fmt.Errorf("API error: code=%d, message=%s", ocrResp.Header.Code, ocrResp.Header.Message)
	}
	return &ocrResp, nil
}

// RecognizeAuto：给我原始图片字节 + GPU 分类（如 "mix0"、"cam.xxx"、"atlas.xxx"）即可。
// 1) 大于 7.5MB 自动压到 ~2MB；2) 自动探测 jpg/png 等编码；3) 自动从分类得到 language(ASECode)。
func (c *Client) RecognizeAuto(ctx context.Context, raw []byte, category string) (*OcrResponse, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty image")
	}

	// 1) 压缩（>7.5MB 才压）
	img := raw
	if int64(len(raw)) > MaxUncompressedBytes {
		// 优先走你现有的压缩工具
		if compressed, err := compressViaToolkit(raw, TargetCompressedBytes); err == nil && len(compressed) > 0 {
			img = compressed
		} else {
			// 兜底：JPEG 品质递减压缩
			if fallback, err2 := compressJPEGFallback(raw, TargetCompressedBytes); err2 == nil && len(fallback) > 0 {
				img = fallback
			} // 若兜底失败就原图发（API 可能直接拒绝）
		}
	}

	// 2) 编码探测
	enc := detectImageEncoding(img)
	if enc == "" {
		enc = "jpg"
	}

	// 3) language 自动从 GPU 分类推导（取第一个 ASECode）
	lang := pickASELanguageFromCategory(category)
	if lang == "" {
		lang = "ch_en" // 兜底中英
	}

	return c.RecognizeBytes(ctx, img, enc, lang)
}

func encodingFromFilename(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return ""
	}
	// 去掉点
	ext = ext[1:]
	switch ext {
	case "jpg", "jpeg":
		return "jpg"
	case "png":
		return "png"
	case "webp":
		return "webp"
	case "gif":
		return "gif"
	default:
		// 尝试从 mime 推断
		if t := mime.TypeByExtension("." + ext); t != "" {
			if strings.Contains(t, "jpeg") {
				return "jpg"
			}
			if strings.Contains(t, "png") {
				return "png"
			}
		}
		return ""
	}
}

func detectImageEncoding(b []byte) string {
	ct := http.DetectContentType(b)
	switch {
	case strings.Contains(ct, "jpeg"):
		return "jpg"
	case strings.Contains(ct, "png"):
		return "png"
	case strings.Contains(ct, "webp"):
		return "webp"
	case strings.Contains(ct, "gif"):
		return "gif"
	default:
		return ""
	}
}

func compressViaToolkit(src []byte, target int64) ([]byte, error) {
	if len(src) == 0 {
		return nil, fmt.Errorf("empty src")
	}
	return utils.CompressImageBytes(src, target)
}

// 兜底：把任何格式解码为 image.Image，然后按 JPEG 品质递减压到目标大小附近（单位：字节）
func compressJPEGFallback(src []byte, target int64) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	quality := 90
	var out []byte
	for quality >= 40 {
		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("jpeg encode: %w", err)
		}
		out = buf.Bytes()
		if int64(len(out)) <= target {
			break
		}
		quality -= 10
	}
	return out, nil
}

// 用你的 ocrblock.GPUCategoryToLanguages 来挑第一候选的 ASECode
func pickASELanguageFromCategory(category string) string {
	if strings.TrimSpace(category) == "" {
		return ""
	}
	langs := GPUCategoryToLanguages(category)
	if len(langs) == 0 || langs[0] == nil {
		return ""
	}
	return langs[0].ASECode
}

// RecognizeBase64 接收 base64（可为 data URL）并调用 OCR API。
// 优先从 data URL 的 MIME 推断 imgEncoding；否则使用调用方传入的 imgEncoding。
func (c *Client) RecognizeBase64(ctx context.Context, b64, imgEncoding, language string) (*OcrResponse, error) {
	if c.AppID == "" || c.APIKey == "" || c.APISecret == "" {
		return nil, fmt.Errorf("AppID/APIKey/APISecret is not configured")
	}
	if strings.TrimSpace(b64) == "" {
		return nil, fmt.Errorf("empty base64 string")
	}

	// 如果是 data URL，提取 mime 与 payload
	if strings.HasPrefix(b64, "data:") {
		mime, payload, ok := parseDataURL(b64)
		if !ok {
			return nil, fmt.Errorf("invalid data URL")
		}
		b64 = payload

		// 仅当调用方未指定 imgEncoding 时，依据 MIME 推断
		if imgEncoding == "" {
			if enc := mimeToImgEncoding(mime); enc != "" {
				imgEncoding = enc
			}
		}
	}

	// 清理空白字符，避免换行导致解码失败
	cleaned := trimAllSpaces(b64)

	// 尝试多种 base64 变体
	var (
		imgBytes []byte
		err      error
	)
	if imgBytes, err = base64.StdEncoding.DecodeString(cleaned); err != nil {
		// 试试 RawStdEncoding（无 '=' padding）
		if imgBytes, err = base64.RawStdEncoding.DecodeString(cleaned); err != nil {
			// 试试 URL-safe 变体
			if imgBytes, err = base64.URLEncoding.DecodeString(cleaned); err != nil {
				if imgBytes, err = base64.RawURLEncoding.DecodeString(cleaned); err != nil {
					return nil, fmt.Errorf("decode base64 failed: %w", err)
				}
			}
		}
	}

	// 回退一个默认值
	if imgEncoding == "" {
		imgEncoding = "jpg"
	}

	return c.RecognizeBytes(ctx, imgBytes, imgEncoding, language)
}

// parseDataURL 解析 data URL，返回 mime、payload(base64 部分)、是否成功
func parseDataURL(s string) (string, string, bool) {
	// 形如：data:image/png;base64,xxxx
	semi := strings.IndexByte(s, ';')
	comma := strings.IndexByte(s, ',')
	if !strings.HasPrefix(s, "data:") || semi == -1 || comma == -1 || semi > comma {
		return "", "", false
	}
	mime := s[len("data:"):semi]
	meta := s[semi+1 : comma]
	if !strings.EqualFold(meta, "base64") && !strings.HasSuffix(strings.ToLower(meta), ";base64") {
		// 仅支持 base64
		return "", "", false
	}
	payload := s[comma+1:]
	return mime, payload, true
}

// mimeToImgEncoding 将 MIME 映射到 API 需要的 imgEncoding
func mimeToImgEncoding(mime string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/bmp":
		return "bmp"
	case "image/gif":
		return "gif"
	case "image/tiff", "image/tif":
		return "tiff"
	default:
		return ""
	}
}

// trimAllSpaces 去掉字符串中的所有空白字符
func trimAllSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
