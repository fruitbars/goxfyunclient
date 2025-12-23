package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/service/ocr"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	appId     string
	apiKey    string
	apiSecret string
	imagePath string
	category  string
	logLevel  string
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("提示: 未找到 .env 文件。")
	}

	appId = os.Getenv("XFYUN_APP_ID")
	apiKey = os.Getenv("XFYUN_API_KEY")
	apiSecret = os.Getenv("XFYUN_API_SECRET")

	flag.StringVar(&imagePath, "file", "", "指定要识别的图片文件路径")
	flag.StringVar(&category, "cat", "ch_en", "指定识别类型 (例如: general, hm_general_ocr, ...)")
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
		logger.Error("凭证未配置，请在 .env 文件中设置 XFYUN_APP_ID, XFYUN_API_KEY, 和 XFYUN_API_SECRET。")
		os.Exit(1)
	}

	client := ocr.NewClient(appId, apiKey, apiSecret, ocr.WithLogger(logger))

	var imageData []byte
	var err error

	if imagePath == "" {
		logger.Info("未指定图片文件，将使用内置的虚拟图片进行测试")
		return
	} else {
		logger.Info("将使用指定的图片文件", "path", imagePath)
		imageData, err = os.ReadFile(imagePath)
	}

	if err != nil {
		logger.Error("读取图片文件失败", "path", imagePath, "error", err)
		os.Exit(1)
	}

	imageBase64 := base64.StdEncoding.EncodeToString(imageData)
	logger.Debug("图片已编码为 Base64")

	logger.Info("", "category", category)
	resp, err := client.RecognizeBase64(context.Background(), imageBase64, "", category)
	//resp, err := client.Recognize(context.Background(), imageBase64, category)
	if err != nil {
		logger.Error("识别失败", "error", err)
		os.Exit(1)
	}

	logger.Info("识别成功", "sid", resp.Header.Sid)

	decodedText, err := base64.StdEncoding.DecodeString(resp.Payload.OcrOutputText.Text)
	if err != nil {
		logger.Error("Base64 解码识别结果失败", "error", err)
		os.Exit(1)
	}

	fmt.Println("--- 识别结果 ---")
	fmt.Println(string(decodedText))

	os.WriteFile(imagePath+".json", decodedText, 0644)
}
