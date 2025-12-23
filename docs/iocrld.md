# 图像文字识别与版面还原 (`iocrld`) 客户端文档

本文档介绍了如何使用 `goxfyunclient` 中的 `iocrld` 客户端来调用讯飞的图像文字识别与版面还原服务。

该服务能够识别图像中的文字，并返回带有坐标和段落信息的详细排版结果，同时可以根据输入的 JSON 对图像进行处理和标记。

## 1. 讯飞 iocrld API 简介

- **官方文档**: (请参考讯飞开放平台相应服务的最新文档)
- **接口地址**: `https://cn-huabei-1.xf-yun.com/v1/private/s15fc3900`
- **鉴权方式**: 使用 `APPID`, `APIKey`, 和 `APISecret` 进行 HMAC-SHA256 签名认证。

## 2. `iocrld` 客户端使用说明

### 2.1. 初始化客户端

```go
import "goxfyunclient/internal/service/iocrld"
import "net/http"

const (
	APP_ID     = "您的APPID"
	API_SECRET = "您的API_SECRET"
	API_KEY    = "您的API_KEY"
)

func main() {
    // 可以传入自定义的 http.Client，如果传入 nil 则使用默认配置
    client := iocrld.NewClient(APP_ID, API_KEY, API_SECRET, iocrld.RequestURL, nil)
    // ... use client
}
```

### 2.2. 处理图像

使用 `Process` 方法是与该服务交互的核心。

**方法签名**:
```go
func (c *Client) Process(ctx context.Context, trackId, pictureBase64 string, jsonDataRaw json.RawMessage) (*Result, error)
```

- `ctx`: 用于控制请求超时和取消的上下文。
- `trackId`: 链路追踪 ID，可自定义，便于日志查询。
- `pictureBase64`: 待识别图片的 Base64 编码字符串。
- `jsonDataRaw`: 一个 `json.RawMessage` 类型的参数，其内容会被 Base64 编码后放入请求体的 `payload.json.text` 字段。这通常用于传递需要叠加到图片上的结构化数据。
- **返回**:
    - `*Result`: 一个包含处理结果的结构体。
        - `SID`: 讯飞返回的会话 ID。
        - `ImageBase64`: (可选) 服务处理后返回的图片 Base64 字符串。
        - `Text`: 服务返回的文本信息，通常是解码后的 JSON 字符串，包含了详细的文字识别和排版结果。
    - `error`: 如果发生网络错误或 API 返回业务错误 (`header.code != 0`)，则返回错误。

**示例代码**:
```go
import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "encoding/base64"
    "goxfyunclient/internal/service/iocrld"
)

// ... (client initialization)

// 1. 读取图片并进行 Base64 编码
picData, err := os.ReadFile("path/to/your/image.jpg")
if err != nil {
    log.Fatalf("读取图片失败: %v", err)
}
picBase64 := base64.StdEncoding.EncodeToString(picData)

// 2. 准备要传入的 JSON 数据 (示例)
// 这部分内容会根据业务需求变化，这里用一个简单的例子
jsonData := map[string]string{"process_instruction": "highlight_titles"}
rawJson, _ := json.Marshal(jsonData)

// 3. 调用 Process 方法
ctx := context.Background()
trackId := "my-unique-request-id-001"
result, err := client.Process(ctx, trackId, picBase64, rawJson)
if err != nil {
    log.Fatalf("处理失败: %v", err)
}

fmt.Printf("处理成功, SID: %s\n", result.SID)
fmt.Printf("返回文本内容: %s\n", result.Text)
// 返回的图片可以保存查看
// if result.ImageBase64 != "" { ... }
```
