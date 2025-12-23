package models

// --- Structs for JSON data ---

type RequestData struct {
	Header    RequestHeader    `json:"header"`
	Parameter RequestParameter `json:"parameter"`
	Payload   RequestPayload   `json:"payload"`
}

type RequestHeader struct {
	AppID  string `json:"app_id"`
	UID    string `json:"uid"`
	Status int    `json:"status"`
}

// ---- Parameter 相关的结构体 ----

// CnenResult 定义了 parameter.cnen.result 的结构
type CnenResult struct {
	Encoding string `json:"encoding"`
	Compress string `json:"compress"`
	Format   string `json:"format"`
}

// CnenParameter 定义了 parameter.cnen 的结构
type CnenParameter struct {
	Outfmt string     `json:"outfmt"`
	Result CnenResult `json:"result"`
}

// RequestParameter 定义了请求的 parameter 部分
type RequestParameter struct {
	Cnen CnenParameter `json:"cnen"`
}

// ---- Payload 相关的结构体 ----

// PayloadRequest 定义了 payload.request 的结构
type PayloadRequest struct {
	Encoding string `json:"encoding"`
	Compress string `json:"compress"`
	Format   string `json:"format"`
	Status   int    `json:"status"`
	Text     string `json:"text"` // base64 编码的文本
}

// RequestPayload 定义了请求的 payload 部分
type RequestPayload struct {
	Request PayloadRequest `json:"request"`
}
