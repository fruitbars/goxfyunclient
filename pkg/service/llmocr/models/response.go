package models

// ResponseBody 定义了从API接收的响应体结构
type ResponseBody struct {
	Header  ResponseHeader  `json:"header"`
	Payload ResponsePayload `json:"payload"`
}

// ResponseHeader 定义了响应头信息
type ResponseHeader struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	SID     string `json:"sid"`
}

// ResponsePayload 包含了OCR识别的结果
type ResponsePayload struct {
	Result OCRResult `json:"result"`
}

// OCRResult 包含了最终解码前的文本
type OCRResult struct {
	Text string `json:"text"`
}
