# 机器翻译 (`translate`) 客户端文档

本文档介绍了如何使用 `goxfyunclient` 中的 `translate` 客户端来调用讯飞的机器翻译服务。

## 1. 讯飞机器翻译 API 简介

- **官方文档**: (请参考讯飞开放平台“机器翻译”服务的最新文档)
- **接口地址**: `https://itrans.xf-yun.com/v1/its`
- **鉴权方式**: 使用 `APPID`, `APIKey`, 和 `APISecret` 进行 HMAC-SHA256 签名认证。

## 2. `translate` 客户端使用说明

### 2.1. 初始化客户端

```go
import "goxfyunclient/internal/service/translate"

const (
	APP_ID     = "您的APPID"
	API_SECRET = "您的API_SECRET"
	API_KEY    = "您的API_KEY"
)

func main() {
    // resID 是可选的热词资源 ID，如果不需要则传入空字符串 ""
    resID := "" 
    client := translate.NewClient(APP_ID, API_KEY, API_SECRET, resID)
    // ... use client
}
```

### 2.2. 执行翻译

使用 `Translate` 方法可以执行文本翻译。

**方法签名**:
```go
func (c *Client) Translate(text, from, to string) (string, error)
```

- `text`: 需要翻译的源文本。
- `from`: 源语种代码，如 `"cn"` (中文)。
- `to`: 目标语种代码，如 `"en"` (英文)。
- **返回**: 翻译后的文本字符串和可能发生的错误。

**语种代码参考**:
| 代码 | 语言 |
|---|---|
| `cn` | 中文 |
| `en` | 英文 |
| `ja` | 日语 |
| `ko` | 韩语 |
| `fr` | 法语 |
| `ru` | 俄语 |
| `es` | 西班牙语 |
| `...`| (更多请参考官方文档) |

**示例代码**:
```go
import (
    "fmt"
    "log"
    "goxfyunclient/internal/service/translate"
)

// ... (client initialization)

sourceText := "讯飞开放平台"
translatedText, err := client.Translate(sourceText, "cn", "en")
if err != nil {
    log.Fatalf("翻译失败: %v", err)
}

fmt.Printf("原文: %s\n", sourceText)
fmt.Printf("译文: %s\n", translatedText)
// 示例输出:
// 原文: 讯飞开放平台
// 译文: iFLYTEK Open Platform
```
