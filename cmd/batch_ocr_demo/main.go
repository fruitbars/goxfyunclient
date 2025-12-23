package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fruitbars/goxfyunclient/pkg/service/ocr"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

var (
	appId     string
	apiKey    string
	apiSecret string
	inputDir  string
	outputDir string
	category  string
	logLevel  string
	workers   int
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("提示: 未找到 .env 文件。")
	}

	appId = os.Getenv("XFYUN_APP_ID")
	apiKey = os.Getenv("XFYUN_API_KEY")
	apiSecret = os.Getenv("XFYUN_API_SECRET")

	flag.StringVar(&inputDir, "input", "", "指定输入目录路径（包含要识别的图片文件）")
	flag.StringVar(&outputDir, "output", "", "指定输出目录路径（用于保存识别结果）")
	flag.StringVar(&category, "cat", "ch_en", "指定识别类型 (例如: general, hm_general_ocr, ...)")
	flag.StringVar(&logLevel, "level", "info", "设置日志级别 (debug, info, warn, error)")
	flag.IntVar(&workers, "workers", 3, "并发工作线程数")
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

	if inputDir == "" {
		logger.Error("必须指定输入目录路径")
		flag.Usage()
		os.Exit(1)
	}

	if outputDir == "" {
		logger.Error("必须指定输出目录路径")
		flag.Usage()
		os.Exit(1)
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("创建输出目录失败", "path", outputDir, "error", err)
		os.Exit(1)
	}

	// 获取输入目录中的所有图片文件
	imageFiles, err := findImageFiles(inputDir)
	if err != nil {
		logger.Error("读取输入目录失败", "path", inputDir, "error", err)
		os.Exit(1)
	}

	if len(imageFiles) == 0 {
		logger.Info("输入目录中没有找到图片文件", "path", inputDir)
		return
	}

	logger.Info("找到图片文件", "count", len(imageFiles), "input", inputDir, "output", outputDir)

	client := ocr.NewClient(appId, apiKey, apiSecret, ocr.WithLogger(logger))

	// 创建并发处理的工作池
	processImages(logger, client, imageFiles, outputDir, workers)

	logger.Info("批量处理完成", "total", len(imageFiles))
}

// findImageFiles 查找目录中的所有图片文件
func findImageFiles(dir string) ([]string, error) {
	var imageFiles []string

	// 支持的图片格式
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".bmp":  true,
		".gif":  true,
		".tiff": true,
		".webp": true,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if imageExts[ext] {
				imageFiles = append(imageFiles, path)
			}
		}
		return nil
	})

	return imageFiles, err
}

// processImages 并发处理图片文件
func processImages(logger *slog.Logger, client *ocr.Client, imageFiles []string, outputDir string, workers int) {
	// 创建工作通道
	jobs := make(chan string, len(imageFiles))
	results := make(chan processResult, len(imageFiles))

	// 启动工作线程
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(i, &wg, jobs, results, logger, client, outputDir)
	}

	// 发送任务
	for _, file := range imageFiles {
		jobs <- file
	}
	close(jobs)

	// 等待所有工作线程完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	successCount := 0
	failCount := 0

	for result := range results {
		if result.success {
			successCount++
			logger.Info("处理成功", "file", result.filename)
		} else {
			failCount++
			logger.Error("处理失败", "file", result.filename, "error", result.err)
		}
	}

	logger.Info("处理统计", "success", successCount, "failed", failCount, "total", len(imageFiles))
}

// processResult 处理结果
type processResult struct {
	filename string
	success  bool
	err      error
}

// worker 工作线程
func worker(id int, wg *sync.WaitGroup, jobs <-chan string, results chan<- processResult,
	logger *slog.Logger, client *ocr.Client, outputDir string) {
	defer wg.Done()

	for file := range jobs {
		logger.Debug("工作线程开始处理", "worker", id, "file", file)

		err := processSingleImage(logger, client, file, outputDir)
		if err != nil {
			results <- processResult{
				filename: file,
				success:  false,
				err:      err,
			}
		} else {
			results <- processResult{
				filename: file,
				success:  true,
			}
		}
	}
}

// processSingleImage 处理单个图片文件
func processSingleImage(logger *slog.Logger, client *ocr.Client, imagePath string, outputDir string) error {
	// 读取图片文件
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("读取图片文件失败: %w", err)
	}

	// 编码为 Base64
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	// 调用 OCR 识别
	resp, err := client.RecognizeBase64(context.Background(), imageBase64, "", category)
	if err != nil {
		return fmt.Errorf("OCR 识别失败: %w", err)
	}

	// 解码识别结果
	decodedText, err := base64.StdEncoding.DecodeString(resp.Payload.OcrOutputText.Text)
	if err != nil {
		return fmt.Errorf("Base64 解码识别结果失败: %w", err)
	}

	// 生成输出文件名（保持原文件名，只改变扩展名）
	baseName := filepath.Base(imagePath)

	outputName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outputName += "_" + resp.Header.Sid + ".json"

	outputPath := filepath.Join(outputDir, outputName)

	// 保存结果到文件
	if err := os.WriteFile(outputPath, decodedText, 0644); err != nil {
		return fmt.Errorf("保存结果文件失败: %w", err)
	}

	logger.Debug("结果已保存", "input", imagePath, "output", outputPath)
	return nil
}
