package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/service/llmocr"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var (
	appId     string
	apiKey    string
	apiSecret string
	imagePath string
	logLevel  string
)

func init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("提示: 未找到 .env 文件，将使用常量中定义的凭证。")
	}

	appId = os.Getenv("XFYUN_APP_ID")
	apiKey = os.Getenv("XFYUN_API_KEY")
	apiSecret = os.Getenv("XFYUN_API_SECRET")

	// 定义命令行参数
	flag.StringVar(&imagePath, "file", "", "指定要识别的图片文件路径")
	flag.StringVar(&logLevel, "level", "info", "设置日志级别 (debug, info, warn, error)")
	flag.Parse()
}

func main() {
	// --- 1. 初始化日志 ---
	var level slog.Level
	switch logLevel {
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
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if appId == "" || apiKey == "" || apiSecret == "" {
		logger.Error("凭证未配置，请在 .env 文件中或代码中设置 APP_ID, API_KEY, 和 API_SECRET。")
		os.Exit(1)
	}

	var localImagePath string

	// 如果未通过命令行指定文件，则创建并使用虚拟图片
	if imagePath == "" {
		localImagePath = "test_llmocr.jpg"
		logger.Info("未指定图片文件，将使用内置的虚拟图片进行测试。")
		if err := createDummyImage(localImagePath); err != nil {
			logger.Error("无法创建虚拟图片", "error", err)
			os.Exit(1)
		}
		defer os.Remove(localImagePath) // 确保程序结束时删除临时文件
	} else {
		localImagePath = imagePath
		logger.Info("将使用指定的图片文件", "path", localImagePath)
	}

	// 创建 llmocr 客户端，并注入 logger
	client := llmocr.NewClient(appId, apiKey, apiSecret, llmocr.WithLogger(logger))

	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 调用OCR服务
	uid := "demo-user-llmocr"
	resultText, err := client.RecognizeFile(ctx, localImagePath, uid)
	if err != nil {
		logger.Error("OCR 识别失败", "error", err)
		os.Exit(1)
	}

	// 尝试将结果解析为JSON并美化输出
	var resultJSON map[string]interface{}
	if err := json.Unmarshal([]byte(resultText), &resultJSON); err == nil {
		prettyJSON, _ := json.MarshalIndent(resultJSON, "", "  ")
		fmt.Printf("\n--- OCR 结果 (JSON) ---\n%s\n", string(prettyJSON))
	} else {
		fmt.Printf("\n--- OCR 结果 (Text) ---\n%s\n", resultText)
	}
}

// createDummyImage 创建一个虚拟的图片文件用于测试
func createDummyImage(filePath string) error {
	// 这是一个1x1像素的白色JPG图像的Base64编码
	base64Image := "/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAIBAQIBAQICAgICAgICAwUDAwMDAwYEBAMFBwYHBwcGBwcICQsJCAgKCAcHCg0KCgsMDAwMBwkODw0MDgsMDAz/2wBDAQICAgMDAwYDAwYMCAcIDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAz/wAARCAABAAEDASIAAhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwD8wKKKK/9k="
	data, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return fmt.Errorf("无法解码虚拟图片: %w", err)
	}
	return os.WriteFile(filePath, data, 0644)
}
