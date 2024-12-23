package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Aran404/Forwarder/api/database"
	"github.com/Aran404/Forwarder/api/solana"
	"github.com/Aran404/Forwarder/api/types"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/gorilla/websocket"
)

const (
	DeadlineContext = time.Second * 30
)

func handleError(w http.ResponseWriter, r *http.Request, status int, reason error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := types.ErrorResponse{
		Success: false,
		Status:  status,
		Message: types.GetProperError(reason),
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// UpgradeWS upgrades the connection to a websocket and runs KeepAlive operations
func (c Client) UpgradeWS(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(DeadlineContext))
	conn.SetWriteDeadline(time.Now().Add(DeadlineContext))
	conn.SetReadLimit(2 << 19)

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(DeadlineContext))
	})

	conn.SetPingHandler(func(string) error {
		return conn.SetWriteDeadline(time.Now().Add(DeadlineContext))
	})

	go func() {
		ticker := time.NewTicker(DeadlineContext)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(DeadlineContext)); err != nil {
					log.Println("Error sending Ping:", err)
					return
				}
			}
		}
	}()

	return conn, nil
}

func (c *Client) Listen() {
	c.http.Post("/payment/create", c.CreatePayment)
	http.ListenAndServe(":3443", c.http)
}

func NewClient(ctx context.Context) *Client {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		httprate.LimitByRealIP(100, time.Minute),
	)

	upgrader := &websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		Error:            handleError,
	}

	return &Client{
		upgrader: upgrader,
		http:     r,
		sol:      solana.NewClient(ctx),
		db:       database.NewConn(ctx),
	}
}

func (c *Client) Close(ctx context.Context) {
	c.db.Close(ctx)
	c.upgrader = nil 
	c.http = nil
}