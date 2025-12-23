package models

type ResponseHeader struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	SID     string `json:"sid"`
}

type ResponseJSONPayload struct {
	Encoding string `json:"encoding"`
	Compress string `json:"compress"`
	Format   string `json:"format"`
	Text     string `json:"text"`
}

type ResponseImagePayload struct {
	Encoding string `json:"encoding"`
	Image    string `json:"image"`
}

type ResponsePayload struct {
	JSON  ResponseJSONPayload  `json:"json"`
	Image ResponseImagePayload `json:"image"`
}

type Response struct {
	Header  ResponseHeader  `json:"header"`
	Payload ResponsePayload `json:"payload"`
}
