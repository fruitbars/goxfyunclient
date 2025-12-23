package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/service/translate"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	appId      string
	apiKey     string
	apiSecret  string
	filePath   string
	sourceLang string
	targetLang string
	logLevel   string
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("提示: 未找到 .env 文件。")
	}

	appId = os.Getenv("XFYUN_APP_ID")
	apiKey = os.Getenv("XFYUN_API_KEY")
	apiSecret = os.Getenv("XFYUN_API_SECRET")

	flag.StringVar(&filePath, "file", "", "包含待翻译文本的文件路径")
	flag.StringVar(&sourceLang, "from", "cn", "源语种 (例如: en, cn)")
	flag.StringVar(&targetLang, "to", "en", "目标语种 (例如: en, cn)")
	flag.StringVar(&logLevel, "level", "info", "设置日志级别 (debug, info, warn, error)")
	flag.Parse()
}

func main() {
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if appId == "" || apiKey == "" || apiSecret == "" {
		logger.Error("凭证未配置, 请在 .env 文件中设置 XFYUN_APP_ID, XFYUN_API_KEY, 和 XFYUN_API_SECRET。")
		os.Exit(1)
	}

	client := translate.NewClient(appId, apiKey, apiSecret, translate.WithLogger(logger))

	textToTranslate, err := getTextToTranslate()
	if err != nil {
		logger.Error("获取待翻译文本失败", "error", err)
		os.Exit(1)
	}

	logger.Info("开始翻译", "from", sourceLang, "to", targetLang, "text_snippet", getSnippet(textToTranslate, 50))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	translatedText, err := client.Translate(ctx, textToTranslate, sourceLang, targetLang)
	if err != nil {
		logger.Error("翻译失败", "error", err)
		os.Exit(1)
	}

	logger.Info("翻译成功")
	fmt.Println("--- 翻译结果 ---")
	fmt.Println(translatedText)
}

func getTextToTranslate() (string, error) {
	var textToTranslate string

	if filePath != "" {
		slog.Info("将从文件中读取文本进行翻译", "file", filePath, "from", sourceLang, "to", targetLang)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取文件 %s 失败: %w", filePath, err)
		}
		textToTranslate = string(content)
	} else if len(flag.Args()) > 0 {
		textToTranslate = strings.Join(flag.Args(), " ")
		slog.Info("将翻译命令行传入的文本", "text", textToTranslate, "from", sourceLang, "to", targetLang)
	} else {
		textToTranslate = "讯飞开放平台"
		slog.Info("未指定文件或文本，将使用默认示例文本", "text", textToTranslate, "from", sourceLang, "to", targetLang)
	}

	return textToTranslate, nil
}

func getSnippet(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length] + "..."
}
