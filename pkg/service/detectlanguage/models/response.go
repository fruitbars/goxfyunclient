package models

// ASELanguageDetectResponse 是 API 返回的最外层结构体
type ASELanguageDetectResponse struct {
	Header  ResponseHeader  `json:"header"`
	Payload ResponsePayload `json:"payload"`
}

// ResponseHeader 定义了响应的 header 部分
type ResponseHeader struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
}

// ResultPayload 定义了 payload.result 的结构
// 其中的 Text 字段是 Base64 编码的、包含最终结果的 JSON 字符串
type ResultPayload struct {
	Encoding string `json:"encoding"`
	Compress string `json:"compress"`
	Format   string `json:"format"`
	Text     string `json:"text"`
}

// ResponsePayload 定义了响应的 payload 部分
type ResponsePayload struct {
	Result ResultPayload `json:"result"`
}

// ---- 以下是解析 Base64 编码的 Text 字段后得到的内部 JSON 结构 ----

// ASELanguageDetectTranResult 是内部 JSON 的顶层结构
type ASELanguageDetectTranResult struct {
	TransResult []TransResultItem `json:"trans_result"`
}

// TransResultItem 定义了 trans_result 数组中每个元素的结构
type TransResultItem struct {
	Src      string `json:"src"`
	LanProbs string `json:"lan_probs"`
}
