package http_helper

import (
	"encoding/json"
	"net/http"
)

func DecodeJSON(r *http.Request, payload interface{}) error {
	return json.NewDecoder(r.Body).Decode(payload)
}

func EncodeJSON(w http.ResponseWriter, status int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

type ResponseBody struct {
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func ResponseError(w http.ResponseWriter, status int, err string, message string) {
	EncodeJSON(w, status, &ResponseBody{
		Error:   err,
		Message: message,
	})
}

func ResponseSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	EncodeJSON(w, status, &ResponseBody{
		Message: message,
		Data:    data,
	})
}
