# 讯飞机器翻译云端接口协议 v1

本文档详细描述了调用讯飞机器翻译 (`its`) 服务的云端 RESTful API 协议。

## 1. 接口概览

- **HTTP Method**: `POST`
- **Endpoint**: `https://itrans.xf-yun.com/v1/its`
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
        "status": 3,
        "res_id": "热词资源ID"
    },
    "parameter": {
        "its": {
            "from": "cn",
            "to": "en",
            "result": {}
        }
    },
    "payload": {
        "input_data": {
            "encoding": "utf8",
            "status": 3,
            "text": "待翻译文本的Base64编码字符串"
        }
    }
}
```

### 3.2. 字段详解

- **header**:
    - `app_id` (string, required): 您的应用 ID。
    - `res_id` (string, optional): 热词资源 ID，用于启用定制化术语翻译。
- **parameter.its**:
    - `from` (string, required): 源语种代码，如 `"cn"`。
    - `to` (string, required): 目标语种代码，如 `"en"`。
- **payload.input_data**:
    - `text` (string, required): 待翻译的 UTF-8 文本内容，需要进行 **Base64 编码**。

## 4. 响应体（Response Body）

### 4.1. 成功响应示例

```json
{
    "header": {
        "code": 0,
        "message": "Success",
        "sid": "its0002111c@dx18b956a9b40f410111"
    },
    "payload": {
        "result": {
            "text": "BASE64_ENCODED_TRANSLATED_TEXT_STRING"
        }
    }
}
```

### 4.2. 字段详解

- **header**:
    - `code` (int): 错误码。`0` 表示成功。
    - `message` (string): 结果描述。
- **payload.result**:
    - `text` (string): **Base64 编码后** 的翻译结果文本。需要解码才能获取可读的译文。

### 4.3. `text` 解码后的内容示例
```json
{
    "trans_result": {
        "dst": "iFLYTEK Open Platform",
        "src": "讯飞开放平台"
    },
    "from": "cn",
    "to": "en"
}
```
通常，我们只需要关心 `trans_result.dst` 字段来获取最终的译文。

### 4.4. 失败响应示例
```json
{
    "header": {
        "code": 10106,
        "message": "Invalid authorization"
    }
}
```
