# 讯飞图像文字识别与版面还原云端接口协议 v1

本文档详细描述了调用讯飞图像文字识别与版面还原 (`iocrld`) 服务的云端 RESTful API 协议。

## 1. 接口概览

- **HTTP Method**: `POST`
- **Endpoint**: `https://cn-huabei-1.xf-yun.com/v1/private/s15fc3900`
- **Content-Type**: `application/json`

## 2. 鉴权机制

接口通过在 URL 查询参数中加入 HMAC-SHA256 签名进行鉴权。鉴权流程与本文档库中其他讯飞 HTTP API 服务一致，使用 `api_key` 字段进行签名。

- **`Authorization` 字段格式**: `api_key="...", algorithm="hmac-sha256", ...`

*详细签名步骤请参考 `detectlanguage_api_protocol.md` 或 `llmocr_api_protocol.md` 中的鉴权章节。*

## 3. 请求体（Request Body）

请求体为一个 JSON 对象，同时包含图像数据和结构化 JSON 数据。

### 3.1. 结构说明

```json
{
    "header": {
        "app_id": "您的APPID",
        "request_id": "请求追踪ID",
        "status": 3
    },
    "parameter": {
        "iocrld": {
            "json": {
                "encoding": "utf8",
                "compress": "raw",
                "format": "json"
            },
            "image": {
                "encoding": "jpg"
            }
        }
    },
    "payload": {
        "json": {
            "encoding": "utf8",
            "compress": "raw",
            "format": "json",
            "status": 3,
            "text": "业务JSON数据的Base64编码字符串"
        },
        "image": {
            "encoding": "jpg",
            "status": 3,
            "image": "待识别图片的Base64编码字符串"
        }
    }
}
```

### 3.2. 字段详解

- **header**:
    - `app_id` (string, required): 您的应用 ID。
    - `request_id` (string, optional): 请求的追踪 ID，建议使用 UUID 或其他唯一标识。
- **payload.json**:
    - `text` (string, required): **Base64 编码后** 的业务 JSON 字符串。服务会解析这个 JSON 来执行特定操作，例如在图片上标记区域。
- **payload.image**:
    - `image` (string, required): 待识别图片的 **Base64 编码** 字符串。

## 4. 响应体（Response Body）

### 4.1. 成功响应示例

```json
{
    "header": {
        "code": 0,
        "message": "Success",
        "sid": "icr0002111c@dx18b956a9b40f410111"
    },
    "payload": {
        "json": {
            "compress": "raw",
            "encoding": "utf8",
            "format": "json",
            "status": 3,
            "text": "BASE64_ENCODED_RESULT_JSON_STRING"
        },
        "image": {
            "compress": "raw",
            "encoding": "jpg",
            "format": "jpg",
            "status": 3,
            "image": "BASE64_ENCODED_RESULT_IMAGE_STRING"
        }
    }
}
```

### 4.2. 字段详解

- **header**:
    - `code` (int): 错误码。`0` 表示成功。
    - `message` (string): 结果描述。
    - `sid` (string): 本次会话的唯一 ID。
- **payload.json**:
    - `text` (string): **Base64 编码后** 的 JSON 结果字符串。解码后是结构化的文字识别和排版信息。
- **payload.image**:
    - `image` (string): **Base64 编码后** 的结果图片。这张图片可能包含了根据输入 JSON 处理后的标记。

### 4.3. 失败响应示例

```json
{
    "header": {
        "code": 10106,
        "message": "Invalid authorization",
        "sid": "icr0002160d@dx18b956a9b6b7410111"
    }
}
```
