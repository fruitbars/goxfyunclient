package models

// RequestBody 定义了发送到API的完整请求体结构
type RequestBody struct {
	Header    Header    `json:"header"`
	Parameter Parameter `json:"parameter"`
	Payload   Payload   `json:"payload"`
}

// Header 定义了请求头信息
type Header struct {
	AppID  string `json:"app_id"`
	UID    string `json:"uid,omitempty"` // omitempty 表示如果字段为空则在JSON中忽略
	Status int    `json:"status"`
}

// Parameter 定义了OCR服务的参数
type Parameter struct {
	OCR OCRParams `json:"ocr"`
}

// OCRParams 定义了具体的OCR处理参数
type OCRParams struct {
	ResultOption string      `json:"result_option"`
	ResultFormat string      `json:"result_format"`
	OutputType   string      `json:"output_type"`
	Result       ResultParam `json:"result"`
}

// ResultParam 定义了结果的编码和格式
type ResultParam struct {
	Encoding string `json:"encoding"`
	Compress string `json:"compress"`
	Format   string `json:"format"`
}

// Payload 包含了需要识别的图像数据
type Payload struct {
	Image ImagePayload `json:"image"`
}

// ImagePayload 定义了图像的具体信息
type ImagePayload struct {
	Encoding string `json:"encoding"`
	Image    string `json:"image"`
	Status   int    `json:"status"`
}
