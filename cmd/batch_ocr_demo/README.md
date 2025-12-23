# 批量OCR识别工具
## 使用说明

### 编译和运行

```bash
go build -o batch_ocr
```

### 命令行参数

```bash
./batch_ocr -input /path/to/input/dir -output /path/to/output/dir [选项]
```

### 参数说明

- `-input`: **必需**，输入目录路径，包含要识别的图片文件
- `-output`: **必需**，输出目录路径，用于保存识别结果
- `-cat`: 识别类型，默认为 "ch_en"
- `-level`: 日志级别，默认为 "info"
- `-workers`: 并发工作线程数，默认为 3

### 使用示例

```bash
# 基本用法
./batch_ocr -input ./images -output ./results

# 指定识别类型和并发数
./batch_ocr -input ./images -output ./results -cat general -workers 5

# 启用调试日志
./batch_ocr -input ./images -output ./results -level debug
```

## 主要特性

1. **批量处理**: 自动扫描输入目录中的所有图片文件
2. **并发处理**: 支持多线程并发处理，提高效率
3. **结果保存**: 将每个图片的识别结果保存为单独的文本文件
4. **错误处理**: 单个文件处理失败不会影响其他文件
5. **进度统计**: 显示处理成功和失败的文件数量
6. **日志记录**: 详细的日志输出，便于调试和监控

## 支持的文件格式

- .jpg, .jpeg
- .png
- .bmp
- .gif
- .tiff
- .webp

输出文件将与输入文件同名，但扩展名改为 `.txt`。