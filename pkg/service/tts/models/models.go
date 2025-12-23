package models

// requestPayload defines the structure for the request sent to the TTS service.
type RequestPayload struct {
	Common   RequestCommon   `json:"common"`
	Business RequestBusiness `json:"business"`
	Data     RequestData     `json:"data"`
}

type RequestCommon struct {
	AppID string `json:"app_id"`
}

type RequestBusiness struct {
	AUE string `json:"aue"`           // Audio encoding
	AUF string `json:"auf"`           // Audio format
	VCN string `json:"vcn"`           // Voice name
	TTE string `json:"tte"`           // Text encoding
	SFL *int   `json:"sfl,omitempty"` // Stream flag for lame
}

type RequestData struct {
	Status int    `json:"status"` // 0 for first frame, 1 for middle, 2 for last
	Text   string `json:"text"`   // Base64 encoded text
}

// serverResponse defines the structure for the response received from the TTS service.
type ServerResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	SID     string `json:"sid"`
	Data    struct {
		Audio  string `json:"audio"`  // Base64 encoded audio data
		Status int    `json:"status"` // 2 indicates the final result
		CED    string `json:"ced"`
	} `json:"data"`
}
