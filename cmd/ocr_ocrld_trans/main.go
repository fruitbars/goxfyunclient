package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/fruitbars/goxfyunclient/pkg/service/ocr"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"os"
)

var XFYUN_APP_ID = "your_app_id"
var XFYUN_API_KEY = ""
var XFYUN_API_SECRET = ""

var (
	imagePath string
	category  string
	logLevel  string
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("提示: 未找到 .env 文件。")
	}

	XFYUN_APP_ID = os.Getenv("XFYUN_APP_ID")
	XFYUN_API_KEY = os.Getenv("XFYUN_API_KEY")
	XFYUN_API_SECRET = os.Getenv("XFYUN_API_SECRET")

	flag.StringVar(&imagePath, "file", "", "指定要识别的图片文件路径")
	flag.StringVar(&category, "cat", "ch_en", "指定识别类型 (例如: general, hm_general_ocr, ...)")
	flag.StringVar(&logLevel, "level", "info", "设置日志级别 (debug, info, warn, error)")
	flag.Parse()
}

// RunOCR 执行 OCR 识别，传入图片路径与类别，返回识别文本
func RunOCR(imagePath, category, logLevel string) (sid, text string, err error) {

	client := ocr.NewClient(XFYUN_APP_ID, XFYUN_API_KEY, XFYUN_API_SECRET)

	if imagePath == "" {
		return "", "", errors.New("imagePath 不能为空")
	}

	// 读取文件
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", "", fmt.Errorf("读取图片失败: %w", err)
	}

	// Base64 编码
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	// 识别
	resp, err := client.RecognizeBase64(context.Background(), imageBase64, "", category)
	if err != nil {
		return "", "", fmt.Errorf("OCR 识别失败: %w", err)
	}

	// Base64 解码结果
	decodedText, err := base64.StdEncoding.DecodeString(resp.Payload.OcrOutputText.Text)
	if err != nil {
		return "", "", fmt.Errorf("识别结果 Base64 解码失败: %w", err)
	}

	return resp.Header.Sid, string(decodedText), nil

}

func main() {
	sid, text, err := RunOCR(imagePath, category, "debug")
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(sid)
	log.Println(text)
}
