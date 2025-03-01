package http

import (
	"encoding/json"
	"net/http"
)

type Message struct {
	Message string `json:"message"`
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
	}
}

func jsonMessage(w http.ResponseWriter, code int, body string) {
	msg := Message{Message: body}

	w.WriteHeader(code)
	jsonResponse(w, msg)
}

func jsonOK(w http.ResponseWriter, body string) {
	jsonMessage(w, http.StatusOK, body)
}

func jsonBadRequest(w http.ResponseWriter, body string) {
	jsonMessage(w, http.StatusBadRequest, body)
}

func jsonUnauthorized(w http.ResponseWriter, body string) {
	jsonMessage(w, http.StatusUnauthorized, body)
}

func jsonForbidden(w http.ResponseWriter, body string) {
	jsonMessage(w, http.StatusForbidden, body)
}

func jsonInternalError(w http.ResponseWriter, body string) {
	jsonMessage(w, http.StatusInternalServerError, body)
}
