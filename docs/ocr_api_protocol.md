# 讯飞通用文字识别云端接口协议 v1

本文档详细描述了调用讯飞通用文字识别 (`ocr`) 服务的云端 RESTful API 协议。

## 1. 接口概览

- **HTTP Method**: `POST`
- **Endpoint**: `https://cn-east-1.api.xf-yun.com/v1/ocr`
- **Content-Type**: `application/json`

## 2. 鉴权机制

接口通过在 URL 查询参数中加入 HMAC-SHA256 签名进行鉴权。鉴权流程与其他讯飞 HTTP API 服务一致，使用 `api_key` 字段进行签名。

- **`Authorization` 字段格式**: `api_key="...", algorithm="hmac-sha256", ...`

*详细签名步骤请参考 `detectlanguage_api_protocol.md` 或 `llmocr_api_protocol.md` 中的鉴权章节。*

## 3. 请求体（Request Body）

### 3.1. 结构说明

```json
{
    "header": {
        "app_id": "您的APPID",
        "status": 3
    },
    "parameter": {
        "ocr": {
            "language": "cn|en",
            "ocr_output_text": {
                "encoding": "utf8",
                "compress": "raw",
                "format": "json"
            }
        }
    },
    "payload": {
        "image": {
            "encoding": "jpg",
            "image": "待识别图片的Base64编码字符串",
            "status": 3
        }
    }
}
```

### 3.2. 字段详解

- **header**:
    - `app_id` (string, required): 您的应用 ID。
- **parameter.ocr**:
    - `language` (string, required): 语种代码。例如 `cn|en` 表示中英混合，`jap` 表示日语。
    - `ocr_output_text`:
        - `encoding` (string): 结果文本编码，`utf8`。
        - `compress` (string): 结果压缩方式，`raw` 表示不压缩。
        - `format` (string): 结果格式，`json`。
- **payload.image**:
    - `encoding` (string, required): 图片的原始格式，如 `jpg`, `png`。
    - `image` (string, required): 图片内容的 **Base64 编码** 字符串。

## 4. 响应体（Response Body）

### 4.1. 成功响应示例

```json
{
    "header": {
        "code": 0,
        "message": "Success",
        "sid": "ocr28926018@dx18b956a9a9a3410111"
    },
    "payload": {
        "ocr_output_text": {
            "compress": "raw",
            "encoding": "utf8",
            "format": "json",
            "status": 3,
            "text": "BASE64_ENCODED_RESULT_JSON_STRING"
        }
    }
}
```

### 4.2. 字段详解

- **header**:
    - `code` (int): 错误码。`0` 表示成功。
    - `message` (string): 结果描述。
- **payload.ocr_output_text**:
    - `text` (string): **Base64 编码后** 的 JSON 结果字符串。解码后是结构化的文字识别结果。

### 4.3. `text` 解码后的内容示例
```json
{
    "pages": [
        {
            "lines": [
                {
                    "words": [
                        {
                            "content": "讯飞开放平台"
                        }
                    ]
                }
            ]
        }
    ]
}
```
内部结构包含了页面 (`pages`)、行 (`lines`)、字 (`words`) 等层级和坐标信息。

### 4.4. 失败响应示例
```json
{
    "header": {
        "code": 10106,
        "message": "Invalid authorization"
    }
}
```
