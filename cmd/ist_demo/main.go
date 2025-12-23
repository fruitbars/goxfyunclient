package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/service/ist"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	appId      string
	secretKey  string
	audioPath  string
	lang       string
	useSpeaker string
	logLevel   string
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("提示: 未找到 .env 文件。")
	}

	appId = os.Getenv("XFYUN_APP_ID")
	secretKey = os.Getenv("XFYUN_SECRET_KEY")

	flag.StringVar(&audioPath, "file", "", "指定要转写的音频文件路径")
	flag.StringVar(&lang, "lang", "cn", "指定语种 (例如: cn, en)")
	flag.StringVar(&useSpeaker, "speaker", "true", "是否开启说话人分离 (true/false)")
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
		level = slog.LevelInfo // Default to info if unknown
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if appId == "" || secretKey == "" {
		logger.Error("凭证未配置，请在 .env 文件中设置 XFYUN_APP_ID 和 XFYUN_SECRET_KEY。")
		os.Exit(1)
	}

	client := ist.NewClient(appId, secretKey, ist.WithLogger(logger))

	var localAudioPath string
	if audioPath == "" {
		localAudioPath = "test_ist.pcm"
		logger.Info("未指定音频文件，将使用内置的虚拟静音音频进行测试")
		if err := createDummyAudio(localAudioPath); err != nil {
			logger.Error("无法创建虚拟音频", "error", err)
			os.Exit(1)
		}
		defer os.Remove(localAudioPath)
	} else {
		localAudioPath = audioPath
	}
	logger.Info("将使用音频文件", "path", localAudioPath)

	logger.Info("这是一个异步过程，可能需要一些时间，请耐心等待...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var opts []ist.UploadOption
	switch lang {
	case "cn":
		//opts = append(opts, ist.WithLanguage(ist.LanguageMandarin))
	case "en":
		//opts = append(opts, ist.WithLanguage(ist.LanguageEnglish))
	default:
		logger.Warn("不支持的语种，将使用默认设置", "lang", lang)
	}
	if useSpeaker == "true" {
		//opts = append(opts, ist.WithSpeakerDiarization())
		logger.Info("已开启说话人分离")
	}

	result, err := client.Process(ctx, localAudioPath, opts...)
	if err != nil {
		logger.Error("处理失败", "error", err)
		os.Exit(1)
	}

	logger.Info("转写成功", "order_id", result.Content.OrderInfo.OrderId)

	if result.Content.OrderResult != "" {
		outputDir := "output"
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			logger.Error("创建输出目录失败", "path", outputDir, "error", err)
			os.Exit(1)
		}

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, []byte(result.Content.OrderResult), "", "  "); err == nil {
			logger.Info("识别结果", "json", prettyJSON.String())
			outputFilePath := fmt.Sprintf("%s/ist_result_%s.json", outputDir, result.Content.OrderInfo.OrderId)
			if err := os.WriteFile(outputFilePath, prettyJSON.Bytes(), 0644); err != nil {
				logger.Error("保存结果失败", "path", outputFilePath, "error", err)
			} else {
				logger.Info("已将详细结果保存", "path", outputFilePath)
			}
		} else {
			logger.Error("无法格式化结果 JSON", "raw_output", result.Content.OrderResult)
		}
	}
}

// createDummyAudio 创建一个虚拟的 PCM 音频文件用于测试
// 内容是一小段静音
func createDummyAudio(filePath string) error {
	// 16-bit, 16kHz, 单声道 PCM 的静音数据 (约 1 秒)
	// RIFF header (not needed for raw PCM, but some players like it)
	// For raw PCM, we just need the data part.
	silentData := make([]byte, 16000*2) // 1 second of 16-bit audio
	return os.WriteFile(filePath, silentData, 0644)
}
