package server

import (
	"math/big"
	"time"

	"github.com/Aran404/Forwarder/api/database"
	"github.com/Aran404/Forwarder/api/solana"
	"github.com/Aran404/Forwarder/api/types"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var (
	ALLOW_LOCAL_HOST = false

	CryptoDeadline        = time.Minute * 30
	MinForward            = types.Config.Forwarder.MinForward                             // Minimum amount to forward in SOL
	TransactionThreshold  = big.NewFloat(1 - types.Config.Forwarder.TransactionThreshold) // Threshold for transaction values
	IgnoreIotaTxThreshold = big.NewFloat(0.02)                                            // If the amount is 2% or less, ignore the transaction. This is to ignore bots.
)

type Client struct {
	upgrader *websocket.Upgrader
	http     *chi.Mux
	sol      *solana.Client
	db       *database.Connection
}

type PaymentCreateBody struct {
	Amount      float64 `json:"amount"`
	CallbackURI string  `json:"callback_uri"`
}

type PaymentCreateResponse struct {
	Success bool    `json:"success"`
	ID      string  `json:"id"`
	Amount  float64 `json:"amount"`
	Address string  `json:"address"`
	QRCode  string  `json:"qrcode"`
	Expires uint64  `json:"expires"`
}

type WebhookResponse struct {
	Success        bool    `json:"success" bson:"success"`
	ID             string  `json:"id" bson:"id"`
	Error          any     `json:"error" bson:"error"`
	DesiredAmount  float64 `json:"desired_amount" bson:"desired_amount"`
	AmountSent     float64 `json:"amount_sent" bson:"amount_sent"`
	TransactionID  string  `json:"transaction_id" bson:"transaction_id"`
	Address        string  `json:"address" bson:"address"`
	TimeSent       uint64  `json:"time_sent" bson:"time_sent"`
	PercentOfTotal float64 `json:"percent_of_total" bson:"percent_of_total"`
}
