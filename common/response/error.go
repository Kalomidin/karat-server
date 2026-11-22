package response

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	statusCode int
	Errors     string `json:"error"`
}

func NewErrorMessage(message string, status int) *Error {
	return &Error{statusCode: status, Errors: message}
}

func (e Error) Send(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.statusCode)
	return json.NewEncoder(w).Encode(e)
}
