package dto

type ErrorResponse struct {
	Error string `json:"error"`
}

type OKResponse struct {
	OK bool `json:"ok"`
}

type ListPageResponse struct {
	Limit    int  `json:"limit"`
	Offset   int  `json:"offset"`
	Total    int  `json:"total"`
	Returned int  `json:"returned"`
	HasMore  bool `json:"hasMore"`
}

type ListQueryResponse struct {
	Filter string `json:"filter,omitempty"`
	Order  string `json:"order,omitempty"`
	Sort   string `json:"sort,omitempty"`
}

type ListResponseMeta struct {
	Page  ListPageResponse  `json:"page"`
	Query ListQueryResponse `json:"query"`
}

func ErrorPayload(errorMessage string) ErrorResponse {
	return ErrorResponse{Error: errorMessage}
}

func OKPayload() OKResponse {
	return OKResponse{OK: true}
}
