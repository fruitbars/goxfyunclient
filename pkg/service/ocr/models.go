package ocr

// ----------- Request / Response structs (align with official demo) -----------

type requestBody struct {
	Header    reqHeader    `json:"header"`
	Parameter reqParameter `json:"parameter"`
	Payload   reqPayload   `json:"payload"`
}

type reqHeader struct {
	AppID  string `json:"app_id"`
	Status int    `json:"status"` // 3: last frame / once-off
}

type reqParameter struct {
	OCR struct {
		Language      string `json:"language"` // e.g. "cn|en", "en", "cn" (按实际产品支持)
		OcrOutputText struct {
			Encoding string `json:"encoding"` // "utf8"
			Compress string `json:"compress"` // "raw"
			Format   string `json:"format"`   // "json"
		} `json:"ocr_output_text"`
	} `json:"ocr"`
}

type reqPayload struct {
	Image struct {
		Encoding string `json:"encoding"` // "jpg" / "png"
		Image    string `json:"image"`    // base64 encoded image
		Status   int    `json:"status"`   // 3
	} `json:"image"`
}

// OcrResponse matches official demo.
type OcrResponse struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Sid     string `json:"sid"`
	} `json:"header"`
	Payload struct {
		OcrOutputText struct {
			Text string `json:"text"` // base64 encoded text
		} `json:"ocr_output_text"`
	} `json:"payload"`
}
