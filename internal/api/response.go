package api

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Meta contains metadata about the response.
type Meta struct {
	Page  int `json:"page"`
	Total int `json:"total"`
}

func writeJSON(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// Success writes a successful response.
func Success(w http.ResponseWriter, statusCode int, message string, data any) {
	if statusCode < 100 || statusCode > 399 {
		statusCode = http.StatusOK
	}
	writeJSON(w, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Fail writes a failed response.
func Fail(w http.ResponseWriter, status int, message string, err error) {
	resp := Response{
		Success: false,
		Message: message,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	if status < 400 || status > 599 {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, resp)
}
