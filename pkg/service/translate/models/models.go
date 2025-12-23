package models

type RequestBody struct {
	Header    RequestHeader    `json:"header"`
	Parameter RequestParameter `json:"parameter"`
	Payload   RequestPayload   `json:"payload"`
}

type RequestHeader struct {
	AppID  string `json:"app_id"`
	Status int    `json:"status"`
	ResID  string `json:"res_id,omitempty"`
}

type RequestParameter struct {
	ITS RequestParameterITS `json:"its"`
}

type RequestParameterITS struct {
	From   string                 `json:"from"`
	To     string                 `json:"to"`
	Result map[string]interface{} `json:"result"`
}

type RequestPayload struct {
	InputData RequestInputData `json:"input_data"`
}

type RequestInputData struct {
	Encoding string `json:"encoding"`
	Status   int    `json:"status"`
	Text     string `json:"text"`
}

type ResponseBody struct {
	Header  ResponseHeader  `json:"header"`
	Payload ResponsePayload `json:"payload"`
}

type ResponseHeader struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
}

type ResponsePayload struct {
	Result ResponseResult `json:"result"`
}

type ResponseResult struct {
	Text string `json:"text"`
}
