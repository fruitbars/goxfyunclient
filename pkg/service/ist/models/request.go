package models

// UploadResponse defines the structure of the JSON response from the upload endpoint.
type UploadResponse struct {
	Code     string `json:"code"`
	DescInfo string `json:"descInfo"`
	Content  struct {
		OrderID          string `json:"orderId"`
		TaskEstimateTime int    `json:"taskEstimateTime"`
	} `json:"content"`
}

// GetResultResponse defines the structure for the transcription result.
type GetResultResponse struct {
	Code     string `json:"code"`
	DescInfo string `json:"descInfo"`
	Content  struct {
		OrderInfo struct {
			OrderId          string `json:"orderId"`
			FailType         int    `json:"failType"`
			Status           int    `json:"status"`
			OriginalDuration int    `json:"originalDuration"`
			RealDuration     int    `json:"realDuration"`
		} `json:"orderInfo"`
		OrderResult      string `json:"orderResult"`
		taskEstimateTime int    `json:"taskEstimateTime"`
	} `json:"content"`
}
