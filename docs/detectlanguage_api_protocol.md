# 讯飞语种识别云端接口协议 v1

本文档详细描述了调用讯飞语种识别服务的云端 RESTful API 协议。

## 1. 接口概览

- **HTTP Method**: `POST`
- **Endpoint**: `https://cn-huadong-1.xf-yun.com/v1/private/s0ed5898e`
- **Content-Type**: `application/json`

## 2. 鉴权机制

接口通过在 URL 查询参数中加入 HMAC-SHA256 签名进行鉴权。

### 2.1. 签名流程

1.  **构造源字符串**:
    将 `host`, `date`, `request-line` 按如下格式拼接。`date` 必须是 GMT 格式的 HTTP `Date` 头 (`RFC1123`)。

    ```
    host: cn-huadong-1.xf-yun.com
    date: Sun, 21 Sep 2025 11:00:00 GMT
    POST /v1/private/s0ed5898e HTTP/1.1
    ```

2.  **HMAC-SHA256 加密**:
    使用您的 `APISecret` 作为密钥，对源字符串进行 HMAC-SHA256 加密，然后进行 Base64 编码，得到签名 `signature`。

3.  **构造 `Authorization` 字符串**:
    将您的 `APIKey` 和上一步生成的 `signature` 按如下格式拼接。**注意**: 此处使用的是 `api_key` 字段。

    ```
    api_key="YOUR_API_KEY", algorithm="hmac-sha256", headers="host date request-line", signature="YOUR_GENERATED_SIGNATURE"
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
https://cn-huadong-1.xf-yun.com/v1/private/s0ed5898e?authorization=BASE64_ENCODED_AUTH_STRING&date=URL_ENCODED_DATE&host=URL_ENCODED_HOST
```

## 3. 请求体（Request Body）

请求体为 JSON 格式。

### 3.1. 结构说明

```json
{
    "header": {
        "app_id": "您的APPID",
        "uid": "用户唯一标识",
        "status": 3
    },
    "parameter": {
        "cnen": {
            "outfmt": "json",
            "result": {
                "encoding": "utf8",
                "compress": "raw",
                "format": "json"
            }
        }
    },
    "payload": {
        "request": {
            "encoding": "utf8",
            "compress": "raw",
            "format": "plain",
            "status": 3,
            "text": "待识别文本的Base64编码字符串"
        }
    }
}
```

### 3.2. 字段详解

- **header**:
    - `app_id` (string, required): 您的应用 ID。
    - `uid` (string, optional): 用户唯一标识，用于区分不同用户。
    - `status` (int, required): 帧状态，固定为 `3`。
- **parameter.cnen.result**:
    - `encoding` (string): 结果文本编码，`utf8`。
    - `compress` (string): 结果压缩方式，`raw` 表示不压缩。
    - `format` (string): 结果格式，`json`。
- **payload.request**:
    - `text` (string, required): **Base64 编码后** 的待识别 UTF-8 文本内容。

## 4. 响应体（Response Body）

### 4.1. 成功响应示例

响应体 `payload.result.text` 字段的值是经过 Base64 编码的 JSON 字符串。

```json
{
    "header": {
        "code": 0,
        "message": "Success",
        "sid": "ltp9496001d@dx18b956a9b0d6410111"
    },
    "payload": {
        "result": {
            "text": "eyAic3JjIjogIuS4iOS6pOiHquebnyIsICJ0cmFuc19yZXN1bHQiOiBbIHsgImxhbl9wcm9icyI6ICJ7XCJjblwiOiAxfSIgfSBdIH0="
        }
    }
}
```

**`text` 字段解码后的内容**:
```json
{
    "src": "你好世界",
    "trans_result": [
        {
            "lan_probs": "{\"cn\": 1}"
        }
    ]
}
```
其中 `lan_probs` 字段的值是一个 JSON 字符串，表示识别出的语种及置信度。

### 4.2. 字段详解

- **header**:
    - `code` (int): 错误码。`0` 表示成功，非 `0` 表示失败。
    - `message` (string): 结果描述。
    - `sid` (string): 本次会话的唯一 ID。
- **payload.result**:
    - `text` (string): **Base64 编码后** 的 JSON 结果字符串。

### 4.3. 失败响应示例

```json
{
    "header": {
        "code": 10110,
        "message": "invalid request",
        "sid": "ltp94938d15@dx18b956a9afcb410111"
    }
}
```
当 `header.code` 不为 0 时，`payload` 字段可能不存在。
