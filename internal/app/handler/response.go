package handler

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if body != nil {
		json.NewEncoder(w).Encode(body)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	resp := map[string]string{"error": message}
	WriteJSON(w, status, resp)
}

func InvalidRequestBody(w http.ResponseWriter) {
	WriteError(w, http.StatusBadRequest, "invalid request body")
}

func InternalServerError(w http.ResponseWriter) {
	WriteError(w, http.StatusInternalServerError, "internal server error")
}
