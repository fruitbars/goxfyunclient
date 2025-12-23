# 语音合成 (`tts`) 客户端文档

本文档介绍了如何使用 `goxfyunclient` 中的 `tts` 客户端来调用讯飞的在线语音合成（Text-to-Speech）服务。

该服务基于 WebSocket 协议，允许客户端发送文本，并实时接收合成的音频流。

## 1. 讯飞语音合成 API 简介

- **官方文档**: (请参考讯飞开放平台“在线语音合成”服务的最新文档)
- **协议**: WebSocket (`wss`)
- **接口地址**: `wss://tts-api.xfyun.cn/v2/tts`
- **鉴权方式**: 通过在 WebSocket 连接 URL 中加入 HMAC-SHA256 签名进行鉴权。

## 2. `tts` 客户端使用说明

客户端封装了 WebSocket 连接、鉴权、文本发送和音频接收的完整流程。

### 2.1. 初始化客户端

使用讯飞开放平台提供的 `AppID`, `APIKey`, 和 `APISecret` 初始化客户端。

```go
import "goxfyunclient/internal/service/tts"

const (
	APP_ID     = "您的APPID"
	API_SECRET = "您的API_SECRET"
	API_KEY    = "您的API_KEY"
)

func main() {
    client := tts.NewTTSClient(APP_ID, API_KEY, API_SECRET)
    // ... use client
}
```

### 2.2. 使用流程

语音合成的流程分为三步：建立连接、发送文本、接收音频。

#### 步骤 1: 建立连接

在发送任何数据之前，必须先调用 `Connect` 方法建立 WebSocket 连接。完成后务必调用 `Close` 断开连接。

```go
err := client.Connect()
if err != nil {
    log.Fatalf("连接失败: %v", err)
}
defer client.Close()
```

#### 步骤 2: 发送文本

使用 `SendText` 方法发送需要合成的文本。该方法可以多次调用以发送连续的文本片段。

```go
textToSynthesize := "你好，欢迎使用讯飞语音合成服务。"
isLastChunk := true // 如果这是最后一段文本，设为 true

err = client.SendText(textToSynthesize, isLastChunk)
if err != nil {
    log.Fatalf("发送文本失败: %v", err)
}
```

#### 步骤 3: 接收音频

循环调用 `ReceiveAudio` 方法来接收合成的音频数据流。

```go
import "os"

// ... (inside a function after sending text)

// 创建一个文件来保存接收到的音频 (PCM 格式)
audioFile, err := os.Create("output.pcm")
if err != nil {
    log.Fatalf("无法创建文件: %v", err)
}
defer audioFile.Close()

for {
    audioData, isLast, err := client.ReceiveAudio()
    if err != nil {
        log.Fatalf("接收音频失败: %v", err)
    }

    if len(audioData) > 0 {
        _, err := audioFile.Write(audioData)
        if err != nil {
            log.Fatalf("写入文件失败: %v", err)
        }
    }

    if isLast {
        fmt.Println("音频接收完毕。")
        break
    }
}
```

### 2.3. 完整示例

```go
package main

import (
	"fmt"
	"goxfyunclient/internal/service/tts"
	"log"
	"os"
)

const (
    // ... 你的凭证 ...
)

func main() {
    client := tts.NewTTSClient(APP_ID, API_KEY, API_SECRET)

    if err := client.Connect(); err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer client.Close()

    text := "这是一段测试文字，将合成为音频。"
    if err := client.SendText(text, true); err != nil {
        log.Fatalf("发送文本失败: %v", err)
    }
    
    // ... (循环接收音频并写入文件的代码) ...
}
```
