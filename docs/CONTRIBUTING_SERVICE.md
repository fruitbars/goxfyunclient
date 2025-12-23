# 新服务模块贡献指南

本文档为向 `goxfyunclient` 项目贡献新服务模块的开发者提供了一套标准的实现原则、结构规范和开发要点。遵循本指南有助于保持项目代码的一致性、可维护性和高质量。

## 1. 核心原则

1.  **单一职责 (Single Responsibility)**: 每个位于 `internal/service/` 下的包应只专注于对接一个具体的讯飞服务（例如 `tts`、`ocr`）。所有与该服务相关的客户端逻辑、模型定义和选项都应内聚在此包内。

2.  **接口优先 (Interface First)**: 在编写具体实现前，优先考虑客户端对外暴露的公共方法。接口应设计得简洁、稳定且易于理解。

3.  **无状态客户端 (Stateless Client)**: 客户端实例 (`Client`) 应该是并发安全的，并且自身不保存任何单次请求相关的状态（例如 WebSocket 连接）。所有请求所需的数据都应通过方法参数传入。

4.  **依赖注入 (Dependency Injection)**: 客户端的所有外部依赖，如日志记录器 (`*slog.Logger`)、自定义 HTTP 客户端 (`*http.Client`) 或 API 地址 (`Host`)，都必须通过**功能选项模式 (Functional Options Pattern)** 进行注入。

## 2. 标准文件结构

每个新的服务模块 `[servicename]` 都应遵循以下文件结构：

```
internal/service/
└── [servicename]/
    ├── client.go          # 客户端核心逻辑实现
    ├── client_test.go     # 针对 client.go 的单元测试
    ├── models.go          # 服务相关的请求/响应结构体定义
    └── options.go         # (可选) 当服务参数复杂时，用于定义功能选项
```

## 3. 开发要点清单

在实现新服务时，请确保遵循以下清单中的每一项：

### 3.1. 客户端构造 (`client.go`)

-   [ ] **必须**使用功能选项模式实现构造函数 `NewClient`。
-   [ ] `NewClient` 的函数签名应为 `NewClient(appID, credential, ...Option)`，其中 `credential` 是 `apiKey` 或 `secretKey` 等。
-   [ ] **必须**提供 `WithLogger(*slog.Logger) Option` 选项。
-   [ ] **必须**提供 `WithHost(string) Option` 选项，用于覆盖默认的 API 服务地址，这对于测试至关重要。
-   [ ] `NewClient` 内部**必须**为 `Logger` 设置一个默认的静默实现 (`slog.New(slog.NewTextHandler(io.Discard, nil))`)，以确保在不注入日志器时，库本身不会产生任何输出。
-   [ ] 客户端 (`Client`) 结构体应包含 `AppID`、密钥、`Host` 和 `Logger` 等基本字段。

### 3.2. 日志记录 (`client.go`)

-   [ ] 在客户端代码中，**禁止**直接使用 `fmt.Print*` 或 `log.Print*` 函数。
-   [ ] **必须**在所有关键路径上使用注入的 `slog.Logger` 实例进行结构化日志记录，例如：
    -   发起网络请求前 (`logger.Debug`)
    -   收到网络响应后 (`logger.Debug`)
    -   发生错误时 (`logger.Error`)
    -   关键业务流程节点 (`logger.Info`)

### 3.3. 错误处理 (`client.go`)

-   [ ] 所有公共方法返回的 `error` 都应包含足够的上下文信息。推荐使用 `fmt.Errorf("...: %w", err)` 来包装底层错误。
-   [ ] 对于讯飞 API 返回的业务错误（`header.code != 0`），应将其封装为一个自定义的错误类型（例如 `APIError`），并包含错误码、错误信息和 SID 等关键信息。

### 3.4. 上下文 (`client.go`)

-   [ ] 所有对外暴露的、会发起网络调用的方法，其第一个参数**必须**是 `context.Context`。
-   [ ] 在执行 `http.NewRequestWithContext` 或处理 WebSocket 连接时，正确地传递和监听 `context` 的取消事件。

### 3.5. 单元测试 (`client_test.go`)

-   [ ] **必须**为所有对外暴露的公共方法编写单元测试。
-   [ ] 对于 HTTP 服务，**必须**使用 `net/http/httptest` 来创建模拟服务器 (`httptest.NewServer`)。
-   [ ] 对于 WebSocket 服务，**必须**使用 `httptest.NewServer` 结合 `gorilla/websocket.Upgrader` 来模拟 WebSocket 服务器。
-   [ ] 测试用例中，通过 `WithHost` 选项将客户端的请求指向模拟服务器。
-   [ ] 测试用例应覆盖至少两种场景：**成功调用的场景**和**API 返回业务错误的场景**。

### 3.6. 文档

-   [ ] 在 `docs/` 目录下，**必须**创建 `[servicename].md` 文件，详细说明该服务客户端的使用方法，并提供代码示例。
-   [ ] 在 `docs/` 目录下，**必须**创建 `[servicename]_api_protocol.md` 文件，详细分析该服务的云端接口协议，包括鉴权、请求/响应结构等。
-   [ ] 在根目录的 `README.md` 和 `docs/README.md` 中的服务列表中，**必须**添加新服务的条目和链接。

## 4. 代码实现模板 (`client.go`)

以下是一个可供参考和复制的简化版 `client.go` 模板，它包含了上述要点的大部分内容。

```go
package myservice

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	defaultHost = "https://example.xf-yun.com/v1/api"
)

// Client for MyService.
type Client struct {
	HostURL    string
	AppID      string
	APIKey     string
	APISecret  string
	Logger     *slog.Logger
	HTTPClient *http.Client
}

// Option is a function that configures a Client.
type Option func(*Client)

// WithHost sets the host for the client.
func WithHost(host string) Option {
	return func(c *Client) {
		if host != "" {
			c.HostURL = host
		}
	}
}

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.Logger = logger
		}
	}
}

// NewClient creates a new MyService client.
func NewClient(appID, apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		HostURL:   defaultHost,
		AppID:     appID,
		APIKey:    apiKey,
		APISecret: apiSecret,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// DoSomething is a public method that calls the service.
func (c *Client) DoSomething(ctx context.Context, text string) (string, error) {
    // ... 1. 鉴权和构建请求 ...
    // ... 2. 使用 c.HTTPClient 发送请求 ...
    // ... 3. 使用 c.Logger 记录日志 ...
    // ... 4. 解析响应并处理错误 ...
    return "result", nil
}
```
