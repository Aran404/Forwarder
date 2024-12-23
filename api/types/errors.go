package types

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

var (
	// Solana Errors
	ErrInvalidStatus        = errors.New("invalid status")
	ErrNoMetadata           = errors.New("no metadata")
	ErrTransactionOverboard = errors.New("transaction has gone overboard")

	// Database Errors
	ErrMustBePointer   = errors.New("must be a pointer")
	ErrNotFound        = errors.New("no matches found")
	ErrFilterCollision = errors.New("collision on filter")

	// HTTP Errors
	ErrNotJSON            = errors.New("not json")
	ErrInvalidCallbackURI = errors.New("invalid callback uri")
	ErrInvalidAmount      = errors.New("invalid amount")
	ErrTransactionSlipped = errors.New("transaction has slipped")

	ProperErrors = map[error]string{
		ErrInvalidStatus:        "Invalid confirmation status.",
		ErrNoMetadata:           "No metadata in transaction.",
		ErrTransactionOverboard: "Transaction has gone overboard, retry with bonded transactions.",
		ErrNotFound:             "No matches found in database.",
		ErrFilterCollision:      "Collision on filter query.",
		ErrNotJSON:              "Request contains invalid JSON. Please use application/json.",
		ErrInvalidCallbackURI:   "Invalid callback uri.",
		ErrInvalidAmount:        "Invalid amount to forward. Please provide a higher amount.",
		ErrTransactionSlipped:   "Transaction has slipped threshold, user has not sent enough funds.",
	}
)

type ErrorResponse struct {
	Success bool   `json:"success"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func GetProperError(err error) string {
	if err == nil {
		return ""
	}
	if v, ok := ProperErrors[err]; ok {
		return v
	}
	return err.Error()
}

func HandleError(w http.ResponseWriter, status int, reason error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := ErrorResponse{
		Success: false,
		Status:  status,
		Message: GetProperError(reason),
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func BadRequest(w http.ResponseWriter, reason error) {
	HandleError(w, http.StatusBadRequest, reason)
}

func InternalServerError(w http.ResponseWriter, reason error) {
	HandleError(w, http.StatusInternalServerError, reason)
}

func GInternalServerError(w http.ResponseWriter) {
	HandleError(w, http.StatusInternalServerError, errors.New("internal server error"))
}

func NotFound(w http.ResponseWriter, reason error) {
	HandleError(w, http.StatusNotFound, reason)
}

func Unauthorized(w http.ResponseWriter, reason error) {
	HandleError(w, http.StatusUnauthorized, reason)
}

func Forbidden(w http.ResponseWriter, reason error) {
	HandleError(w, http.StatusForbidden, reason)
}
