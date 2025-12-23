# 讯飞语音合成（TTS）云端接口协议 v2

本文档详细描述了调用讯飞在线语音合成（TTS）服务的云端 WebSocket API 协议。

## 1. 接口概览

- **协议**: WebSocket (`wss`)
- **Endpoint**: `wss://tts-api.xfyun.cn/v2/tts`
- **请求行**: `GET /v2/tts HTTP/1.1`

## 2. 鉴权与连接

通过构建一个包含鉴权信息的 URL 来建立 WebSocket 连接。

### 2.1. 鉴权 URL 构建

1.  **构造源字符串**:
    将 `host`, `date`, `request-line` 按如下格式拼接。`date` 必须是 GMT 格式的 HTTP `Date` 头 (`RFC1123`)。`host` 与请求地址的 `host` 一致。

    **源字符串示例**:
    ```
    host: tts-api.xfyun.cn
    date: Thu, 01 Aug 2019 01:53:21 GMT
    GET /v2/tts HTTP/1.1
    ```

2.  **HMAC-SHA256 加密与 Base64 编码**:
    使用 `APISecret` 作为密钥对源字符串进行加密，然后对加密结果进行 Base64 编码，得到 `signature`。
    将 `APIKey`, `signature` 等信息拼接为 `authorization` 字符串，最后对 `authorization` 字符串整体进行 Base64 编码。详细规则请参考官方文档。


3.  **构建最终 URL**:
    将鉴权信息作为查询参数拼接到 Endpoint 后面。

    **最终 URL 示例**:
    ```
    wss://tts-api.xfyun.cn/v2/tts?authorization=BASE64_ENCODED_AUTH_STRING&date=URL_ENCODED_DATE&host=tts-api.xfyun.cn
    ```

## 3. 通信协议

连接建立后，客户端和服务器通过 JSON 格式的帧进行通信。

### 3.1. 客户端发送帧

客户端向服务器发送包含待合成文本的 JSON 对象。

#### 结构说明
```json
{
    "common": {
        "app_id": "您的APPID"
    },
    "business": {
        "aue": "raw",
        "auf": "audio/L16;rate=16000",
        "vcn": "x4_yezi",
        "tte": "UTF8"
    },
    "data": {
        "status": 2,
        "text": "待合成文本的Base64编码字符串"
    }
}
```
#### 字段详解
- **common.app_id** (string, required): 您的应用 ID。
- **business**:
    - `aue` (string, required): 音频编码格式。`raw` 表示未压缩的 PCM。`lame` 表示 MP3。
    - `auf` (string, optional): 音频采样率等参数。`aue`为`raw`时，`audio/L16;rate=16000` 表示 16k 采样率的 PCM。
    - `vcn` (string, required): 发音人。例如 `x4_yezi`。
    - `tte` (string, required): 文本编码，`UTF8`。
- **data**:
    - `status` (int, required): 帧状态。对于在线合成，固定为 `2`。
    - `text` (string, required): **Base64 编码后** 的待合成 UTF-8 文本。

### 3.2. 服务器响应帧

服务器会返回一个或多个包含音频数据的 JSON 对象。

#### 结构说明
```json
{
    "code": 0,
    "message": "success",
    "sid": "tts000f57a9@dx18b956a9b9b0410111",
    "data": {
        "audio": "BASE64_ENCODED_AUDIO_CHUNK",
        "status": 2,
        "ced": "1,2;3,5"
    }
}
```
#### 字段详解
- **code** (int): 错误码。`0` 表示成功。
- **data**:
    - `audio` (string): **Base64 编码后** 的音频数据片段 (PCM)。
    - `status` (int): 帧状态。`1` 表示还有后续音频，`2` 表示这是最后一帧音频。
    - `ced` (string): 合成字符的边界信息。

### 3.3. 通信流程
1. 客户端建立 WebSocket 连接。
2. 客户端发送一个包含所有文本的请求帧 (`status`=2)。
3. 服务器返回一个或多个响应帧。
4. 当客户端收到 `data.status` 为 `2` 的响应帧时，表示合成结束。
5. 客户端关闭 WebSocket 连接。
