package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/service/iocrld"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	appId       string
	apiKey      string
	apiSecret   string
	imagePath   string
	jsonPayload string
	logLevel    string
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("提示: 未找到 .env 文件。")
	}

	appId = os.Getenv("XFYUN_APP_ID")
	apiKey = os.Getenv("XFYUN_API_KEY")
	apiSecret = os.Getenv("XFYUN_API_SECRET")

	flag.StringVar(&imagePath, "file", "", "指定要识别的图片文件路径")
	flag.StringVar(&jsonPayload, "payload", `{"param":{"extract_title":true}}`, "指定业务处理的 JSON 字符串")
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
		logger.Error("凭证未配置")
		os.Exit(1)
	}

	// 创建 iocrld 客户端
	client := iocrld.NewClient(appId, apiKey, apiSecret, iocrld.WithLogger(logger))

	var localImagePath string
	if imagePath == "" {
		localImagePath = "test_iocrld.jpg"
		logger.Info("未指定图片文件，将使用内置的虚拟图片进行测试")
		if err := createDummyImage(localImagePath); err != nil {
			logger.Error("无法创建虚拟图片", "error", err)
			os.Exit(1)
		}
		defer os.Remove(localImagePath)
	} else {
		localImagePath = imagePath
	}
	logger.Info("参数信息", "image_path", localImagePath, "payload", jsonPayload)

	picData, err := os.ReadFile(localImagePath)
	if err != nil {
		logger.Error("读取图片失败", "path", localImagePath, "error", err)
		os.Exit(1)
	}
	picBase64 := base64.StdEncoding.EncodeToString(picData)

	logger.Info("开始调用图像识别与版面还原服务")
	ctx := context.Background()
	trackId := "demo-request-iocrld-001"
	var params map[string]interface{}
	if strings.TrimSpace(jsonPayload) != "" {
		if err := json.Unmarshal([]byte(jsonPayload), &params); err != nil {
			logger.Error("payload 不是合法 JSON", "error", err, "payload", jsonPayload)
			os.Exit(1)
		}
	} else {
		params = map[string]interface{}{} // 没给就用空对象
	}

	result, err := client.Process(ctx, trackId, picBase64, params)
	if err != nil {
		logger.Error("处理失败", "error", err)
		os.Exit(1)
	}

	logger.Info("处理成功", "sid", result.Header.SID)

	if result.Payload.JSON.Text != "" {
		logger.Info("识别文本结果 (格式化后)")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, []byte(result.Payload.JSON.Text), "", "  "); err == nil {
			logger.Info("识别文本结果", "text", prettyJSON.String())
		} else {
			logger.Info("无法格式化 JSON，原始输出", "text", result.Payload.JSON.Text)
		}
	}

	if result.Payload.Image.Image != "" {
		outputDir := "output"
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			logger.Error("创建输出目录失败", "path", outputDir, "error", err)
			os.Exit(1)
		}

		decodedImage, err := base64.StdEncoding.DecodeString(result.Payload.Image.Image)
		if err != nil {
			logger.Error("解码返回的图片失败", "error", err)
		} else {
			outputImagePath := fmt.Sprintf("%s/result_iocrld_%d.jpg", outputDir, time.Now().Unix())
			err := os.WriteFile(outputImagePath, decodedImage, 0644)
			if err != nil {
				logger.Error("保存返回的图片失败", "path", outputImagePath, "error", err)
			} else {
				logger.Info("已将处理后的图片保存", "path", outputImagePath)
			}
		}
	}
}

// createDummyImage 创建一个虚拟的图片文件用于测试 (同 llmocr_demo)
func createDummyImage(filePath string) error {
	base64Image := "/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAIBAQIBAQICAgICAgICAwUDAwMDAwYEBAMFBwYHBwcGBwcICQsJCAgKCAcHCg0KCgsMDAwMBwkODw0MDgsMDAz/2wBDAQICAgMDAwYDAwYMCAcIDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAz/wAARCAABAAEDASIAAhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwD8wKKKK/9k="
	data, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return fmt.Errorf("无法解码虚拟图片: %w", err)
	}
	return os.WriteFile(filePath, data, 0644)
}
