package models

// models.go

// InnerResultPayload 定义了被 Base64 编码在 ResponseBody.Payload.Result.Text 中的 JSON 结构
type EngineResultPayload struct {
	TransResult EngineTransResult `json:"trans_result"`
}

// InnerTransResult 包含最终的翻译结果
type EngineTransResult struct {
	Dst string `json:"dst"` // 目标语言翻译结果
	Src string `json:"src"` // 源语言文本
}
