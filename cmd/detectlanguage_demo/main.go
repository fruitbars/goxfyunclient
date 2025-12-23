package main

import (
	"bufio"
	"flag"
	"github.com/fruitbars/goxfyunclient/pkg/service/detectlanguage"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	appId     string
	apiKey    string
	apiSecret string
	filePath  string
	logLevel  string
)

func init() {
	_ = godotenv.Load() // 忽略 .env 文件不存在的情况

	appId = os.Getenv("XFYUN_APP_ID")
	apiKey = os.Getenv("XFYUN_API_KEY")
	apiSecret = os.Getenv("XFYUN_API_SECRET")

	flag.StringVar(&filePath, "file", "", "指定包含待识别文本的文件路径")
	flag.StringVar(&logLevel, "level", "info", "日志级别：debug|info|warn|error")
	flag.Parse()
}

func main() {
	// 初始化日志
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if appId == "" || apiKey == "" || apiSecret == "" {
		logger.Error("凭证未配置，请在 .env 或环境变量中设置 XFYUN_APP_ID, XFYUN_API_KEY, XFYUN_API_SECRET")
		os.Exit(1)
	}

	if filePath == "" {
		logger.Error("请使用 --file 参数指定待识别文本文件路径")
		os.Exit(1)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("无法打开文件", "path", filePath, "error", err)
		os.Exit(1)
	}
	defer file.Close()

	// 创建客户端
	client := detectlanguage.NewClient(appId, apiKey, apiSecret, detectlanguage.WithLogger(logger))
	logger.Info("开始进行语种识别", "file", filePath)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		lineNum++
		if text == "" {
			continue
		}

		result, err := client.Detect(text)
		if err != nil {
			logger.Error("文本识别失败", "line", lineNum, "text", truncate(text, 50), "error", err)
			continue
		}

		logger.Info("识别成功",
			"line", lineNum,
			"text", truncate(text, 50),
			"result", result,
		)
	}
	if err := scanner.Err(); err != nil {
		logger.Error("读取文件出错", "error", err)
		os.Exit(1)
	}

	logger.Info("识别完成")
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "..."
}
