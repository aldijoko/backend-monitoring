package dto

type EdgeFormPayload struct {
	Code      string `json:"code" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
}
