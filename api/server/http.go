package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Aran404/Forwarder/api/types"
)

func ParseJSON(r *http.Request, v interface{}) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if r.Header.Get("Content-Type") != "application/json" {
		return types.ErrNotJSON
	}

	return json.Unmarshal(data, v)
}

func SendJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func SendWebhook(uri string, v interface{}) {
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(v); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}

	_, err := http.Post(uri, "application/json", body)
	if err != nil {
		log.Printf("Error sending webhook: %v", err)
	}
}

func (c *Client) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var body *PaymentCreateBody
	if err := ParseJSON(r, &body); err != nil {
		types.BadRequest(w, err)
		return
	}

	if !strings.HasPrefix(body.CallbackURI, "https://") {
		body.CallbackURI = "https://" + body.CallbackURI
	}

	if !strings.HasPrefix(body.CallbackURI, "http://") {
		body.CallbackURI = "http://" + body.CallbackURI
	}

	if strings.Contains(body.CallbackURI, "localhost") {
		types.BadRequest(w, types.ErrInvalidCallbackURI)
		return
	}

	if body.Amount < MinForward || body.Amount <= 0 {
		types.BadRequest(w, types.ErrInvalidAmount)
		return
	}

	c.HandleCreatePayment(w, r, body)
}
