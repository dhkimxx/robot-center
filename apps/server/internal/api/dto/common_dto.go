package dto

type ErrorResponse struct {
	Error string `json:"error"`
}

type OKResponse struct {
	OK bool `json:"ok"`
}

func ErrorPayload(errorMessage string) ErrorResponse {
	return ErrorResponse{Error: errorMessage}
}

func OKPayload() OKResponse {
	return OKResponse{OK: true}
}
