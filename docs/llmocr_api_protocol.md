# 讯飞 LLM OCR 云端接口协议 v1

本文档详细描述了调用讯飞星火大模型图像识别（LLM OCR）服务的云端 RESTful API 协议。

## 1. 接口概览

- **HTTP Method**: `POST`
- **Endpoint**: `https://cbm01.cn-huabei-1.xf-yun.com/v1/private/se75ocrbm`
- **Content-Type**: `application/json`

## 2. 鉴权机制

接口通过在 URL 查询参数中加入 HMAC-SHA256 签名进行鉴权。

### 2.1. 签名流程

1.  **构造源字符串**:
    将 `host`, `date`, `request-line` 按如下格式拼接。`date` 必须是 GMT 格式的 HTTP `Date` 头。

    ```
    host: cbm01.cn-huabei-1.xf-yun.com
    date: Sun, 21 Sep 2025 10:00:00 GMT
    POST /v1/private/se75ocrbm HTTP/1.1
    ```

2.  **HMAC-SHA256 加密**:
    使用您的 `APISecret` 作为密钥，对源字符串进行 HMAC-SHA256 加密，然后进行 Base64 编码，得到签名 `signature`。

3.  **构造 `Authorization` 字符串**:
    将您的 `APIKey` 和上一步生成的 `signature` 按如下格式拼接。

    ```
    hmac username="YOUR_API_KEY", algorithm="hmac-sha256", headers="host date request-line", signature="YOUR_GENERATED_SIGNATURE"
    ```

4.  **Base64 编码 `Authorization`**:
    对上一步生成的 `Authorization` 字符串整体进行 Base64 编码。

### 2.2. 构建最终请求 URL

将鉴权信息作为查询参数拼接到 Endpoint 后面。

- `authorization`: 步骤 2.1.4 中生成的 Base64 编码字符串。
- `date`: GMT 格式的时间字符串 (需 URL Encode)。
- `host`: API 的主机名 (需 URL Encode)。

**最终 URL 示例**:
```
https://cbm01.cn-huabei-1.xf-yun.com/v1/private/se75ocrbm?authorization=BASE64_ENCODED_AUTH_STRING&date=URL_ENCODED_DATE&host=URL_ENCODED_HOST
```

## 3. 请求体（Request Body）

请求体为 JSON 格式。

### 3.1. 结构说明

```json
{
    "header": {
        "app_id": "您的APPID",
        "uid": "用户唯一标识"
    },
    "parameter": {
        "ocr": {
            "result_option": "normal",
            "result_format": "json",
            "output_type": "one_shot",
            "result": {
                "encoding": "utf8",
                "compress": "raw",
                "format": "plain"
            }
        }
    },
    "payload": {
        "image": {
            "encoding": "图片格式",
            "image": "图片的Base64编码字符串",
            "status": 3
        }
    }
}
```

### 3.2. 字段详解

- **header**:
    - `app_id` (string, required): 您的应用 ID。
    - `uid` (string, optional): 用户唯一标识，用于区分不同用户，可自定义。
- **parameter.ocr**:
    - `result_option` (string): 结果选项，`normal` 为普通模式。
    - `result_format` (string): 结果格式，推荐使用 `json` 获取结构化数据。
    - `output_type` (string): 输出类型，`one_shot` 表示一次性输出全部结果。
    - **result**:
        - `encoding` (string): 结果文本编码，`utf8`。
        - `compress` (string): 结果压缩方式，`raw` 表示不压缩。
        - `format` (string): 结果格式，`plain` 表示纯文本。
- **payload.image**:
    - `encoding` (string, required): 图片的原始格式，如 `jpg`, `png` 等 (不带`.`)。
    - `image` (string, required): 图片内容的 Base64 编码字符串。
    - `status` (int, required): 帧状态，固定为 `3`，表示最后一帧。

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
        "result": {
            "text": "BASE64_ENCODED_RESULT_STRING"
        }
    }
}
```

### 4.2. 字段详解

- **header**:
    - `code` (int): 错误码。`0` 表示成功，非 `0` 表示失败。
    - `message` (string): 结果描述。成功时为 "Success"，失败时为错误信息。
    - `sid` (string): 本次会话的唯一 ID。
- **payload.result**:
    - `text` (string): **Base64 编码后** 的识别结果。解码后即为 `parameter.ocr.result.format` 所指定格式的字符串（例如 JSON 或纯文本）。

### 4.3. 失败响应示例

```json
{
    "header": {
        "code": 10106,
        "message": "Invalid authorization",
        "sid": "ocrfa73240a@dx18b956a9acb5410111"
    }
}
```
当 `header.code` 不为 0 时，`payload` 字段可能不存在。开发者应通过 `code` 和 `message` 来排查问题。
