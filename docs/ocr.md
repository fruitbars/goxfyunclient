# 通用文字识别 (`ocr`) 客户端文档

本文档介绍了如何使用 `goxfyunclient` 中的 `ocr` 客户端来调用讯飞的通用文字识别服务。

该客户端提供了标准和高级两种使用方式，支持自动图像压缩、编码探测和语种推断。

## 1. 讯飞通用文字识别 API 简介

- **官方文档**: (请参考讯飞开放平台“通用文字识别”服务的最新文档)
- **接口地址**: `https://cn-east-1.api.xf-yun.com/v1/ocr`
- **鉴权方式**: 使用 `APPID`, `APIKey`, 和 `APISecret` 进行 HMAC-SHA256 签名认证。

## 2. `ocr` 客户端使用说明

### 2.1. 初始化客户端

```go
import "goxfyunclient/internal/service/ocr"

const (
	APP_ID     = "您的APPID"
	API_SECRET = "您的API_SECRET"
	API_KEY    = "您的API_KEY"
)

func main() {
    client := ocr.NewClient(APP_ID, API_KEY, API_SECRET)
    // ... use client
}
```

### 2.2. 标准识别（从字节流）

使用 `RecognizeBytes` 方法可以从内存中的图片数据进行识别。

**方法签名**:
```go
func (c *Client) RecognizeBytes(ctx context.Context, image []byte, imgEncoding, language string) (*OcrResponse, error)
```

- `ctx`: 上下文。
- `image`: 图片文件的字节内容。
- `imgEncoding`: 图片的编码格式，如 `"jpg"`, `"png"`。
- `language`: 语种代码，如 `"cn|en"` (中英混合)。
- **返回**: 包含完整 API 响应的 `*OcrResponse` 结构体和错误。

**获取识别文本**:
调用 `RecognizeBytes` 成功后，可以方便地从响应对象中提取解码后的文本：
```go
resp, err := client.RecognizeBytes(...)
if err != nil { /* ... */ }

text, err := resp.RecognizedText()
if err != nil { /* ... */ }

fmt.Println(text)
```

### 2.3. 高级识别（推荐）

使用 `RecognizeAuto` 方法提供了更强大和便捷的功能，它会自动处理图片压缩、格式检测和语种推断。

**方法签名**:
```go
func (c *Client) RecognizeAuto(ctx context.Context, raw []byte, category string) (*OcrResponse, error)
```

- `raw`: 原始图片文件的字节内容。
- `category`: 图片的分类信息（例如来自内部图像分类服务），客户端会据此推断 `language` 参数。
- **内部处理流程**:
    1.  **自动压缩**: 如果图片大于 7.5MB，会自动尝试压缩到 2MB 左右。
    2.  **格式检测**: 自动检测图片的编码格式 (jpg, png 等)。
    3.  **语种推断**: 根据输入的 `category` 字符串推断出最合适的语种代码。

**示例代码**:
```go
import (
    "context"
    "fmt"
    "log"
    "os"
    "goxfyunclient/internal/service/ocr"
)

// ... (client initialization)

imageData, err := os.ReadFile("path/to/large_image.jpg")
if err != nil {
    log.Fatalf("读取图片失败: %v", err)
}

// 假设我们通过其他服务得知该图片分类为 "mix0" (中英混合场景)
imageCategory := "mix0" 

ctx := context.Background()
resp, err := client.RecognizeAuto(ctx, imageData, imageCategory)
if err != nil {
    log.Fatalf("自动识别失败: %v", err)
}

text, _ := resp.RecognizedText()
fmt.Println("识别结果:", text)

```
