package utils

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"net/http"
	"os"
	"strings"
)

// CompressImageBytes 压缩图片字节数据到指定大小以下。
// inputData: 输入的图片字节数据。
// maxSize:   最大允许的文件大小(单位: Bytes)。
// 返回: 压缩后的图片字节数据, 错误信息。
func CompressImageBytes(inputData []byte, maxSize int64) ([]byte, error) {
	if int64(len(inputData)) <= maxSize {
		return inputData, nil
	}

	format := detectImagingFormat(inputData)

	img, err := imaging.Decode(bytes.NewReader(inputData))
	if err != nil {
		return nil, fmt.Errorf("无法解码图片数据: %w", err)
	}

	// 定义质量等级和尺寸缩放比例
	qualities := []int{90, 80, 75, 70, 60, 50}
	ratios := []float64{1.0, 0.9, 0.8, 0.7, 0.6, 0.5} // 1.0 表示原始尺寸

	// 【优化策略】使用嵌套循环，对每个尺寸都尝试所有质量等级
	for _, ratio := range ratios {
		var currentImg image.Image
		if ratio == 1.0 {
			currentImg = img
		} else {
			currentImg = resizeByRatio(img, ratio)
		}

		for _, quality := range qualities {
			outputData, err := encodeWithQuality(currentImg, format, quality)
			if err != nil {
				// 编码失败，可能无需继续尝试
				return nil, err
			}
			if int64(len(outputData)) <= maxSize {
				return outputData, nil
			}
		}
	}

	return nil, fmt.Errorf("无法将图片压缩到指定大小(%d bytes)", maxSize)
}

// encodeWithQuality 使用 imaging 包统一编码图片并设置质量。
func encodeWithQuality(img image.Image, format imaging.Format, quality int) ([]byte, error) {
	var buf bytes.Buffer
	// imaging.Encode 可以处理所有格式，imaging.JPEGQuality 只对 JPEG 生效，会被其他格式忽略。
	err := imaging.Encode(&buf, img, format, imaging.JPEGQuality(quality))
	if err != nil {
		return nil, fmt.Errorf("编码图片失败: %w", err)
	}
	return buf.Bytes(), nil
}

// CompressImage 负责文件的读写和调用核心压缩函数。
func CompressImage(inputPath, outputPath string, maxSize int64) error {
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("无法读取输入文件: %w", err)
	}

	outputData, err := CompressImageBytes(inputData, maxSize)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, outputData, 0644)
}

// resizeByRatio 按比例缩放图片。
func resizeByRatio(img image.Image, ratio float64) image.Image {
	// 当 height 参数为 0 时，imaging.Resize 会自动保持宽高比
	width := int(float64(img.Bounds().Dx()) * ratio)
	return imaging.Resize(img, width, 0, imaging.Lanczos)
}

// detectImagingFormat 从字节数据中检测图片格式。
func detectImagingFormat(data []byte) imaging.Format {
	contentType := http.DetectContentType(data)
	switch {
	case strings.Contains(contentType, "jpeg"):
		return imaging.JPEG
	case strings.Contains(contentType, "png"):
		return imaging.PNG
	case strings.Contains(contentType, "gif"):
		return imaging.GIF
	case strings.Contains(contentType, "bmp"):
		return imaging.BMP
	case strings.Contains(contentType, "tiff"):
		return imaging.TIFF
	default:
		// 默认或未知格式，选择 JPEG 作为兜底
		return imaging.JPEG
	}
}
