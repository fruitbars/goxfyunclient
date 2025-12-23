# 讯飞语音转写（IST）云端接口协议 v1

本文档详细描述了调用讯飞语音转写（长时间语音，IST/LFAASR）服务的云端 API 协议。该服务采用异步工作模式。

## 1. 接口概览

- **服务地址**: `https://raasr.xfyun.cn/v2/api`
- **协议流程**:
    1.  **上传音频**: 调用 `/upload` 接口，提交音频文件和转写参数，获取 `orderId`。
    2.  **获取结果**: 调用 `/getResult` 接口，使用 `orderId` 轮询处理进度，直至获取最终转写结果。

## 2. 鉴权机制

两个接口均采用相同的签名鉴权方式，通过 URL 查询参数传递。

### 2.1. 签名 `signa` 生成流程

1.  获取当前时间的 Unix 时间戳（秒），作为字符串 `ts`。
2.  将 `AppID` 和 `ts` 拼接，计算 MD5 值。
    - `baseString = AppID + ts`
    - `md5String = md5(baseString)`
3.  使用 `SecretKey`作为密钥，对上一步的 `md5String` 进行 HMAC-SHA1 加密。
4.  将加密后的二进制结果进行 Base64 编码，即得到最终的 `signa`。

## 3. 接口详情

### 3.1. 上传音频 (`/upload`)

- **HTTP Method**: `POST`
- **Content-Type**: `application/octet-stream`
- **Body**: 音频文件的二进制内容。

#### URL 查询参数

| 参数名 | 类型 | 必选 | 说明 |
|---|---|---|---|
| `appId` | string | 是 | 您的应用 ID |
| `signa` | string | 是 | 根据步骤 2.1 生成的签名 |
| `ts` | string | 是 | 当前 Unix 时间戳（秒） |
| `fileSize`| string | 是 | 原始音频文件大小（字节） |
| `fileName`| string | 是 | 带后缀的文件名 |
| `duration`| string | 是 | 音频时长（秒） |
| `language`| string | 否 | 语种，如 `cn` (中文普通话), `en` (英文) |
| `speaker_diarization` | string | 否 | `1` 表示开启话者分离 |
| `callback_url` | string | 否 | 转写结果回调地址 |
| `...` | ... | 否 | 其他更多转写参数 |

#### 响应体 (Response Body)

**成功示例**:
```json
{
    "code": "000000",
    "descInfo": "success",
    "content": {
        "orderId": "D2D5E63B4C7D4A2C9F8A7E6B5C4D3E2F"
    }
}
```
- `code`: 结果码，`000000` 表示成功。
- `content.orderId`: 订单 ID，用于后续结果查询。

### 3.2. 获取结果 (`/getResult`)

- **HTTP Method**: `POST`

#### URL 查询参数

| 参数名 | 类型 | 必选 | 说明 |
|---|---|---|---|
| `appId` | string | 是 | 您的应用 ID |
| `signa` | string | 是 | 根据步骤 2.1 生成的签名 |
| `ts` | string | 是 | 当前 Unix 时间戳（秒） |
| `orderId` | string | 是 | `/upload` 接口返回的订单 ID |
| `resultType`| string | 否 | 结果格式，`json` 或 `srt`，默认为 `json`|

#### 响应体 (Response Body)

**处理中示例**:
```json
{
    "code": "000000",
    "descInfo": "success",
    "content": {
        "orderInfo": {
            "status": 3,
            "failType": -1
        }
    }
}
```

**成功示例**:
```json
{
    "code": "000000",
    "descInfo": "success",
    "content": {
        "orderId": "D2D5E63B4C7D4A2C9F8A7E6B5C4D3E2F",
        "orderInfo": {
            "status": 4,
            "failType": -1
        },
        "orderResult": "{\"lattice\":[...]}"
    }
}
```
- `content.orderInfo.status`: 订单状态。`3` 表示进行中，`4` 表示已完成，其他值表示失败。
- `content.orderResult`: **字符串形式** 的转写结果。当 `resultType` 为 `json` 时，这是一个需要再次进行 JSON 解析的字符串。

#### 轮询逻辑

客户端应以一定的时间间隔（如 5-10 秒）调用本接口，直到 `content.orderInfo.status` 变为 `4` (成功) 或其他非 `3` 的值 (失败)。
