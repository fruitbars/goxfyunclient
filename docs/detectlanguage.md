# 语种识别 (`detectlanguage`) 客户端文档

本文档介绍了如何使用 `goxfyunclient` 中的 `detectlanguage` 客户端来调用讯飞的语种识别服务。

## 1. 讯飞语种识别 API 简介

讯飞的语种识别服务能够识别文本所属的语言种类。

- **官方文档**: (请参考讯飞开放平台相应服务的最新文档)
- **接口地址**: `https://cn-huadong-1.xf-yun.com/v1/private/s0ed5898e`
- **鉴权方式**: 使用 `APPID`, `APIKey`, 和 `APISecret` 进行 HMAC-SHA256 签名认证，认证信息通过 URL 查询参数传递。

## 2. `detectlanguage` 客户端使用说明

### 2.1. 初始化客户端

首先，你需要从讯飞开放平台获取你的 `APPID`, `APIKey`, 和 `APISecret`，然后使用它们来创建一个新的客户端实例。

```go
import "goxfyunclient/internal/service/detectlanguage"

const (
	APP_ID     = "您的APPID"
	API_SECRET = "您的API_SECRET"
	API_KEY    = "您的API_KEY"
)

func main() {
    client := detectlanguage.NewClient(APP_ID, API_KEY, API_SECRET)
    // ... use client
}
```

### 2.2. 识别语种

使用 `Detect` 方法可以识别输入文本的语种。

**方法签名**:
```go
func (c *Client) Detect(text string) (string, error)
```

- `text`: 需要识别语种的 UTF-8 编码的字符串。
- **返回**: 包含语种及其置信度的 JSON 字符串和可能发生的错误。

**示例代码**:
```go
import (
    "fmt"
    "log"
    "goxfyunclient/internal/service/detectlanguage"
)

// ... (client initialization)

textToDetect := "你好，世界！"
result, err := client.Detect(textToDetect)
if err != nil {
    log.Fatalf("语种识别失败: %v", err)
}

fmt.Printf("识别结果: %s\n", result)
// 示例输出: 识别结果: {"cn": 1}
```
