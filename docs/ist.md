# 语音转写 (`ist`) 客户端文档

本文档介绍了如何使用 `goxfyunclient` 中的 `ist` 客户端来调用讯飞的语音转写（长时间语音）服务。

该服务适用于较长的音频文件，采用异步处理模式：先上传音频文件获取一个订单号，然后通过轮询该订单号来获取最终的转写结果。

## 1. 讯飞语音转写 API 简介

- **官方文档**: (请参考讯飞开放平台“语音转写”服务的最新文档)
- **接口地址**: `https://raasr.xfyun.cn/v2/api`
- **核心流程**:
    1.  POST `/upload`: 上传音频文件，获取 `orderId`。
    2.  POST `/getResult`: 使用 `orderId` 轮询查询处理结果。
- **鉴权方式**: 使用 `AppID` 和 `SecretKey` 生成签名 `signa`，通过 URL 查询参数传递。

## 2. `ist` 客户端使用说明

客户端封装了文件上传、签名生成和结果轮询的完整流程，推荐使用一体化的 `Process` 方法。

### 2.1. 初始化客户端

使用讯飞开放平台提供的 `AppID` 和 `SecretKey` 初始化客户端。

```go
import "goxfyunclient/internal/service/ist"

const (
	APP_ID     = "您的APPID"
	SECRET_KEY = "您的SecretKey" // 注意：此处是 SecretKey，不是 APISecret 或 APIKey
)

func main() {
    client := ist.NewClient(APP_ID, SECRET_KEY)
    // ... use client
}
```

### 2.2. 处理音频文件（推荐）

使用 `Process` 方法可以一步完成音频上传和结果获取的全部流程。

**方法签名**:
```go
func (c *Client) Process(ctx context.Context, filePath string, options ...UploadOption) (*models.GetResultResponse, error)
```

- `ctx`: 用于控制整个处理流程的超时和取消。
- `filePath`: 本地音频文件的路径。
- `options`: (可选) 一系列配置项，用于指定转写参数，例如语种、是否开启说话人分离等。
- **返回**:
    - `*models.GetResultResponse`: 包含完整转写结果的结构体。
    - `error`: 如果处理过程中任何一步失败，则返回错误。

**示例代码**:
```go
import (
    "context"
    "fmt"
    "log"
    "goxfyunclient/internal/service/ist"
)

// ... (client initialization)

ctx := context.Background()
// 使用 WithLanguage 和 WithSpeakerDiarization 等选项来配置转写参数
// 更多选项请参考 options.go 文件
result, err := client.Process(
    ctx,
    "path/to/your/audio.mp3",
    ist.WithLanguage(ist.LanguageMandarin),
    ist.WithSpeakerDiarization(),
)
if err != nil {
    log.Fatalf("处理失败: %v", err)
}

// 成功后，结果在 result.Content.OrderResult 中
// 这是一个JSON字符串，需要进一步解析
fmt.Printf("转写成功, 订单号: %s\n", result.Content.OrderId)
fmt.Printf("原始结果JSON: %s\n", result.Content.OrderResult)
```

### 2.3. 自定义选项 `UploadOption`

可以通过传递不同的 `UploadOption` 来控制转写行为。

- `ist.WithLanguage(lang)`: 设置语种，如 `ist.LanguageMandarin` (普通话)。
- `ist.WithSpeakerDiarization()`: 开启说话人分离（话者分离）。
- `ist.WithCallbackURL("http://...")`: 设置结果回调地址。
- 更多选项请参考 `internal/service/ist/options.go` 文件。

### 2.4. 分步操作（高级）

对于需要更精细控制的场景，可以分别调用 `UploadFile` 和 `GetTranscriptionResult`。

1.  **上传文件**:
    ```go
    orderId, err := client.UploadFile(ctx, "path/to/audio.mp3")
    ```
2.  **获取结果**:
    ```go
    result, err := client.GetTranscriptionResult(ctx, orderId, "json") // resultType: "json" or "srt"
    ```
