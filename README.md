# Go Xunfei (iFlyTek) Client

[![LICENSE](https://img.shields.io/github/license/OWNER/REPOSITORY)](./LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/OWNER/REPOSITORY)](https://goreportcard.com/report/github.com/OWNER/REPOSITORY)
[![GoDoc](https://godoc.org/github.com/OWNER/REPOSITORY?status.svg)](https://godoc.org/github.com/OWNER/REPOSITORY)
[![CI/CD](https://github.com/OWNER/REPOSITORY/actions/workflows/go.yml/badge.svg)](https://github.com/OWNER/REPOSITORY/actions/workflows/go.yml)

`goxfyunclient` 是一个为科大讯飞（iFlyTek）开放平台各项服务编写的 Go 语言客户端库。它旨在简化与讯飞 API 的交互，提供了一套干净、现代且易于使用的接口。

本项目包含了各个服务的客户端实现、高质量的命令行演示程序以及详细的 API 文档。

**请注意：** 请将 README 中的 `OWNER/REPOSITORY` 占位符替换为你的实际 GitHub 用户名和仓库名。

## 📚 目录
- [✨ 功能特性](#-功能特性)
- [🚀 快速开始](#-快速开始)
  - [1. 安装](#1-安装)
  - [2. 配置凭证](#2-配置凭证)
  - [3. 使用示例 (以 LLM OCR 为例)](#3-使用示例-以-llm-ocr-为例)
- [📦 已支持的服务](#-已支持的服务)
- [⚙️ 运行 Demo](#️-运行-demo)
- [📚 文档](#-文档)
- [🤝 贡献](#-贡献)
- [📄 许可证](#-许可证)

## ✨ 功能特性

- **全面的服务支持**: 为多个讯飞 AI 服务提供了客户端封装。
- **现代化 Go 实践**: 使用 `context`、结构化日志（`slog`）和模块化的项目结构。
- **易于使用**: 每个服务都提供了独立的客户端和清晰的调用方法。
- **高质量演示**: `cmd` 目录下为每个服务提供了功能齐全的、可通过命令行参数配置的演示程序。
- **配置简单**: 通过 `.env` 文件集中管理所有 API 凭证，无需硬编码。
- **经过测试**: 包含了单元测试以确保客户端的稳定性和可靠性。
- **详细文档**: `docs` 目录下包含了每个服务的客户端使用说明和云端 API 协议文档。

## 🚀 快速开始

### 1. 安装

首先，确保你的项目中已经初始化了 Go Modules：
```bash
go mod init your-project-name
```

然后，获取本客户端库（请将 `OWNER/REPOSITORY` 替换为实际地址）：
```bash
# 实际使用时，请替换为你的仓库地址
go get github.com/OWNER/REPOSITORY
```

### 2. 配置凭证

在你的项目根目录下创建一个名为 `.env` 的文件，并填入从讯飞开放平台获取的凭证信息。

```dotenv
# .env file
# 讯飞开放平台通用凭证
XFYUN_APP_ID="your_app_id"
XFYUN_API_KEY="your_api_key"
XFYUN_API_SECRET="your_api_secret"

# 特定服务可能需要不同的 key，例如语音转写（ist）
# XFYUN_SECRET_KEY="your_secret_key_for_ist"
```

### 3. 使用示例 (以 LLM OCR 为例)

下面是一个如何在你的代码中使用本库进行大模型 OCR 识别的简单示例。

```go
package main

import (
	"context"
	"fmt"
	"goxfyunclient/internal/service/llmocr"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// 1. 加载 .env 配置
	if err := godotenv.Load(); err != nil {
		log.Fatal("无法加载 .env 文件")
	}
	appId := os.Getenv("XFYUN_APP_ID")
	apiKey := os.Getenv("XFYUN_API_KEY")
	apiSecret := os.Getenv("XFYUN_API_SECRET")
	if appId == "" || apiKey == "" || apiSecret == "" {
		log.Fatal("请确保 .env 文件中已配置讯飞凭证")
	}
	
	// 2. 初始化客户端和日志
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := llmocr.NewClient(appId, apiKey, apiSecret, logger)

	// 3. 调用识别接口
	imagePath := "path/to/your/image.jpg"
	resultText, err := client.RecognizeFile(context.Background(), imagePath, "") // uid 可以为空
	if err != nil {
		logger.Error("识别失败", "error", err)
		return
	}

	// 4. 输出结果
	fmt.Println("--- 识别结果 ---")
	fmt.Println(resultText)
}
```

## 📦 已支持的服务

本库目前支持以下讯飞 AI 服务：

| 服务名称 | 模块 (`internal/service`) | Demo (`cmd`) | 功能描述 |
| :--- | :--- | :--- | :--- |
| **大模型 OCR** | `llmocr` | `llmocr_demo` | 基于大模型的通用文字识别 |
| **语种识别** | `detectlanguage` | `detectlanguage_demo` | 识别文本所属的语言种类 |
| **版面还原 OCR** | `iocrld` | `iocrld_demo` | 识别图片文字并还原其版面布局 |
| **语音转写** | `ist` | `ist_demo` | 长语音文件的异步转写服务 |
| **通用 OCR** | `ocr` | `ocr_demo` | 通用场景的文字识别 |
| **机器翻译** | `translate` | `translate_demo` | 多语种之间的文本翻译 |
| **语音合成** | `tts` | `tts_demo` | 将文本转换为自然流畅的语音 |

## ⚙️ 运行 Demo

`cmd` 目录下为每个服务都提供了一个命令行演示程序。你可以使用 `go run` 来执行它们。

每个 demo 都支持 `-h` 或 `--help` 参数来查看其具体用法。

**通用运行方式:**

1.  确保根目录存在 `.env` 文件并已正确配置。
2.  进入具体的 demo 目录并执行 `go run`。

**示例：运行 `iocrld` 版面还原 Demo**
```bash
# 识别一张图片
go run ./cmd/iocrld_demo/main.go --file "path/to/your/document.png"

# 查看帮助
go run ./cmd/iocrld_demo/main.go --help
```

所有 demo 的输出结果（如图片、音频、JSON 文件）默认会保存在项目根目录下的 `output` 文件夹中。

## 📚 文档

更详细的客户端使用说明和每个服务的云端 API 协议分析，请查阅 `docs` 目录下的相关文档。

- [**服务概览**](./docs/overview.md)
- ... (其他文档链接)

## 🤝 贡献

我们非常欢迎社区的贡献！无论是提交 Issue、发起 Pull Request，还是改进文档，都对项目有很大帮助。

在开始贡献之前，请花点时间阅读我们的贡献指南（即将创建）。

## 📄 许可证

本项目采用 [MIT 许可证](./LICENSE)。
