package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/service/tts"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	appId      string
	apiKey     string
	apiSecret  string
	filePath   string
	outputFile string
	logLevel   string
	vcn        string
	aue        string
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("提示: 未找到 .env 文件。")
	}

	appId = os.Getenv("XFYUN_APP_ID")
	apiKey = os.Getenv("XFYUN_API_KEY")
	apiSecret = os.Getenv("XFYUN_API_SECRET")

	flag.StringVar(&filePath, "file", "", "包含待合成文本的文件路径")
	flag.StringVar(&outputFile, "out", "", "指定输出音频文件路径 (例如: output/tts_result.mp3)")
	flag.StringVar(&logLevel, "level", "debug", "设置日志级别 (debug, info, warn, error)")
	flag.StringVar(&vcn, "vcn", "xiaoyan", "设置发音人 (例如: xiaoyan)")
	flag.StringVar(&aue, "aue", "mp3", "设置音频编码格式 (例如: mp3, raw)")
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
	logger := slog.New(
		slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:     level,
			AddSource: true, // 开启文件名和行号
		}),
	)

	if appId == "" || apiKey == "" || apiSecret == "" {
		logger.Error("凭证未配置, 请在 .env 文件中设置 XFYUN_APP_ID, XFYUN_API_KEY, 和 XFYUN_API_SECRET。")
		os.Exit(1)
	}

	client := tts.NewTTSClient(appId, apiKey, apiSecret, tts.WithLogger(logger))

	textToConvert, err := getTextToConvert()
	if err != nil {
		logger.Error("获取待合成文本失败", "error", err)
		os.Exit(1)
	}

	logger.Info("开始合成", "text_snippet", getSnippet(textToConvert, 50), "vcn", vcn, "aue", aue)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	audioData, err := client.TextToSpeech(ctx, textToConvert, vcn, aue)
	if err != nil {
		logger.Error("语音合成失败", "error", err)
		os.Exit(1)
	}

	if len(audioData) == 0 {
		logger.Warn("合成成功，但未收到任何音频数据。")
		return
	}

	outputFilePath := getOutputFilePath()
	err = saveAudioToFile(audioData, outputFilePath)
	if err != nil {
		logger.Error("保存音频文件失败", "path", outputFilePath, "error", err)
		os.Exit(1)
	}

	logger.Info("语音合成成功", "output_path", outputFilePath, "audio_size_kb", len(audioData)/1024)
}

func getTextToConvert() (string, error) {
	if filePath != "" {
		slog.Info("将从文件中读取文本进行语音合成", "file_path", filePath)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取文件 %s 失败: %w", filePath, err)
		}
		return string(content), nil
	} else if len(flag.Args()) > 0 {
		text := strings.Join(flag.Args(), " ")
		slog.Info("将合成命令行传入的文本", "text", text)
		return text, nil
	} else {
		defaultText := "讯飞语音合成技术，让机器“开口说话”，为您带来流畅、自然、真实的听觉体验。"
		slog.Info("未指定文件或文本，将使用默认示例文本", "default_text", defaultText)
		return defaultText, nil
	}
}

func getOutputFilePath() string {
	if outputFile != "" {
		slog.Info("音频将保存到", "output_file", outputFile)
		return outputFile
	}
	defaultOutputFile := "output/tts_result.mp3"
	slog.Info("未指定输出文件，将使用默认输出文件路径", "default_output_file", defaultOutputFile)
	return defaultOutputFile
}

func saveAudioToFile(data []byte, filePath string) error {
	slog.Info("正在保存音频数据到文件", "file_path", filePath, "data_size_bytes", len(data))
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	audioFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", filePath, err)
	}
	defer audioFile.Close()

	_, err = audioFile.Write(data)
	if err != nil {
		return fmt.Errorf("写入音频数据到文件失败: %w", err)
	}
	slog.Info("音频数据已成功写入文件", "file_path", filePath, "data_size_bytes", len(data))
	return nil
}

func getSnippet(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length] + "..."
}
