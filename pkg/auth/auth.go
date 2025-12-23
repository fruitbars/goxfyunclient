package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SchemeType 定义了支持的讯飞鉴权模式的枚举。
// 使用具名类型让代码更清晰、更安全。
type SchemeType int

const (
	// SchemeTypeAPIKey 使用 api_key 字段进行鉴权。
	// 适用于：TTS、通用 OCR 等服务。
	SchemeTypeAPIKey SchemeType = iota // 默认值为 0

	// SchemeTypeHMAC 使用 hmac username 字段进行鉴权。
	// 适用于：LLM OCR 等服务。
	SchemeTypeHMAC
)

// AssembleAuthURL generates the final request URL with authentication parameters for Xunfei services.
func AssembleAuthURL(requestURL, method, apiKey, apiSecret string) (string, error) {
	u, err := url.Parse(requestURL)
	if err != nil {
		return "", err
	}
	host := u.Host
	path := u.Path
	date := time.Now().UTC().Format(time.RFC1123)

	// 1. Create the signature string
	signatureOrigin := fmt.Sprintf("host: %s\ndate: %s\n%s %s HTTP/1.1", host, date, method, path)

	// 2. Encrypt with HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(signatureOrigin))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// 3. Create the authorization string
	authOrigin := fmt.Sprintf(
		`api_key="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`,
		apiKey, signature,
	)

	// 4. Base64 encode the authorization string
	authorization := base64.StdEncoding.EncodeToString([]byte(authOrigin))

	// 5. Build the final URL
	v := url.Values{}
	v.Add("host", host)
	v.Add("date", date)
	v.Add("authorization", authorization)

	return requestURL + "?" + v.Encode(), nil
}

// AssembleAuthURLWithHostPath generates the final request URL with authentication parameters using provided scheme, host, and path.
func AssembleAuthURLWithHostPath(scheme, host, path, method, apiKey, apiSecret string) (string, error) {

	date := time.Now().UTC().Format(time.RFC1123)
	signatureOrigin := fmt.Sprintf("host: %s\ndate: %s\n%s %s HTTP/1.1", host, date, method, path)
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(signatureOrigin))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	authOrigin := fmt.Sprintf(`api_key="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`, apiKey, signature)
	authorization := base64.StdEncoding.EncodeToString([]byte(authOrigin))

	v := url.Values{}
	v.Add("host", host)
	v.Add("date", date)
	v.Add("authorization", authorization)

	return fmt.Sprintf("%s://%s%s?%s", scheme, host, path, v.Encode()), nil
}

// BuildAuthURL 创建一个经过签名的讯飞服务 URL，支持多种鉴权模式。
// 这个函数是可配置的，旨在替换所有独立的、重复的鉴权函数。
//
// 参数:
//
//	baseURL:    带协议和域名的完整基础 URL (例如 "https://cbm01.cn-huabei-1.xf-yun.com/v1/private/se75ocrbm")
//	method:     HTTP 方法 ("POST", "GET" 等)
//	apiKey:     服务的 API Key
//	apiSecret:  服务的 API Secret
//	schemeType: 鉴权模式，使用本包中定义的 SchemeType 常量
func BuildAuthURL(baseURL, method, apiKey, apiSecret string, schemeType SchemeType) (string, error) {
	// 1. 解析基础 URL 以获取 Host 和 Path (公共逻辑)
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %w", err)
	}

	// 2. 准备签名所需的公共组件 (公共逻辑)
	date := time.Now().UTC().Format(http.TimeFormat)
	signatureOrigin := fmt.Sprintf("host: %s\ndate: %s\n%s %s HTTP/1.1", u.Host, date, method, u.Path)

	// 3. 计算 HMAC-SHA256 签名 (公共逻辑)
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(signatureOrigin))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// 4. 【核心区别】根据传入的 schemeType 选择不同的 authorization 格式
	var authOrigin string
	switch schemeType {
	case SchemeTypeAPIKey:
		authOrigin = fmt.Sprintf(`api_key="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`, apiKey, signature)
	case SchemeTypeHMAC:
		authOrigin = fmt.Sprintf(`hmac username="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`, apiKey, signature)
	default:
		return "", fmt.Errorf("不支持的鉴权模式: %d", schemeType)
	}

	// 5. 将 authorization 字符串进行 Base64 编码 (公共逻辑)
	authorization := base64.StdEncoding.EncodeToString([]byte(authOrigin))

	// 6. 拼接最终的 URL (公共逻辑)
	v := url.Values{}
	v.Add("host", u.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)

	return baseURL + "?" + v.Encode(), nil
}
