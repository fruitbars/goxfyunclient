# LLM OCR 服务客户端文档

本文档介绍了如何使用 `goxfyunclient` 中的 `llmocr` 客户端来调用讯飞星火大模型的图像识别（LLM OCR）服务。

## 1. 讯飞 LLM OCR API 简介

讯飞的 LLM OCR 服务提供了高精度的通用文字识别能力。

- **官方文档**: [讯飞开放平台通用文字识别文档](https://www.xfyun.cn/doc/words/universal-ocr/API.html) (请以官方最新链接为准)
- **接口地址**: `https://cbm01.cn-huabei-1.xf-yun.com/v1/private/se75ocrbm`
- **鉴权方式**: 使用 `APPID`, `APIKey`, 和 `APISecret` 进行 HMAC-SHA256 签名认证。

## 2. `llmocr` 客户端使用说明

### 2.1. 初始化客户端

首先，你需要从讯飞开放平台获取你的 `APPID`, `APIKey`, 和 `APISecret`，然后使用它们来创建一个新的客户端实例。

```go
import "goxfyunclient/internal/service/llmocr"

const (
	APP_ID     = "您的APPID"
	API_SECRET = "您的API_SECRET"
	API_KEY    = "您的API_KEY"
)

func main() {
    client := llmocr.NewClient(APP_ID, API_KEY, API_SECRET)
    // ... use client
}
```

### 2.2. 从文件识别文字

使用 `RecognizeFile` 方法可以从本地图片文件中识别文字。

**方法签名**:
```go
func (c *Client) RecognizeFile(ctx context.Context, imagePath, uid string) (string, error)
```

- `ctx`: 上下文，用于控制请求超时和取消。
- `imagePath`: 本地图片文件的路径。
- `uid`: 用户唯一标识符（可自定义）。
- **返回**: 识别出的文字内容（通常是JSON格式的字符串）和可能发生的错误。

**示例代码**:
```go
// ... (client initialization)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.RecognizeFile(ctx, "path/to/your/image.jpg", "user-id-123")
if err != nil {
    log.Fatalf("识别失败: %v", err)
}
fmt.Println(result)
```

### 2.3. 从字节流识别文字

使用 `RecognizeBytes` 方法可以直接从内存中的图片数据（`[]byte`）进行识别，无需先保存为文件。

**方法签名**:
```go
func (c *Client) RecognizeBytes(ctx context.Context, imageData []byte, imageType, uid string) (string, error)
```

- `imageData`: 图片文件的字节内容。
- `imageType`: 图片的类型（例如："jpg", "png"），不包含 `.`。
- 其他参数同 `RecognizeFile`。

**示例代码**:
```go
// ... (client initialization)
imageData, err := os.ReadFile("path/to/your/image.jpg")
if err != nil {
    log.Fatalf("读取图片失败: %v", err)
}

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.RecognizeBytes(ctx, imageData, "jpg", "user-id-123")
if err != nil {
    log.Fatalf("识别失败: %v", err)
}
fmt.Println(result)
```

## 3. 运行演示程序

项目在 `cmd/llmocr_demo` 目录下提供了一个完整的可运行示例。

1. **修改凭证**: 打开 `cmd/llmocr_demo/main.go` 文件，将文件顶部的 `APP_ID`, `API_KEY`, `API_SECRET` 替换为您自己的凭证。
2. **运行程序**:
   ```bash
   go run cmd/llmocr_demo/main.go
   ```
3. **查看输出**: 程序会创建一个临时的 `test.jpg` 文件，调用识别服务，然后将格式化后的识别结果打印到控制台。
