package models

type Header struct {
	AppID     string  `json:"app_id"`
	UID       string  `json:"uid,omitempty"`
	DID       string  `json:"did,omitempty"`
	IMEI      string  `json:"imei,omitempty"`
	IMSI      string  `json:"imsi,omitempty"`
	MAC       string  `json:"mac,omitempty"`
	NetType   string  `json:"net_type,omitempty"`
	NetISP    string  `json:"net_isp,omitempty"`
	Status    int     `json:"status"`
	RequestID *string `json:"request_id,omitempty"`
	ResID     string  `json:"res_id,omitempty"`
}

type JSONParams struct {
	Encoding string `json:"encoding"`
	Compress string `json:"compress"`
	Format   string `json:"format"`
}

type ImageParams struct {
	Encoding string `json:"encoding"`
}

type IOCRld struct {
	JSON  JSONParams  `json:"json"`
	Image ImageParams `json:"image"`
}

type Parameter struct {
	IOCRld IOCRld `json:"iocrld"`
}

type JSONPayload struct {
	Encoding string `json:"encoding"`
	Compress string `json:"compress"`
	Format   string `json:"format"`
	Status   int    `json:"status"`
	Text     string `json:"text"`
}

type ImagePayload struct {
	Encoding string `json:"encoding"`
	Image    string `json:"image"`
	Status   int    `json:"status"`
}

type Payload struct {
	JSON  JSONPayload  `json:"json"`
	Image ImagePayload `json:"image"`
}

type Request struct {
	Header    Header    `json:"header"`
	Parameter Parameter `json:"parameter"`
	Payload   Payload   `json:"payload"`
}
