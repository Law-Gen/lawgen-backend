package response

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/domain"
)

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func WriteOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

func WriteCreated(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, data)
}

func WriteBadRequest(w http.ResponseWriter, code, message string, details map[string]string) {
	writeJSON(w, http.StatusBadRequest, ErrorBody{Code: code, Message: message, Details: details})
}

func WriteUnauthorized(w http.ResponseWriter, code, message string) {
	writeJSON(w, http.StatusUnauthorized, ErrorBody{Code: code, Message: message})
}

func WriteForbidden(w http.ResponseWriter, code, message string) {
	writeJSON(w, http.StatusForbidden, ErrorBody{Code: code, Message: message})
}

func WriteNotFound(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusNotFound, ErrorBody{Code: domain.ErrNotFound, Message: message})
}

func WriteError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, ErrorBody{Code: domain.ErrServerError, Message: "An unexpected error occurred."})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func DecodeJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

var ErrNotFound = errors.New("not found")