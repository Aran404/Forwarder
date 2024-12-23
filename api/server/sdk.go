package server

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/Aran404/Forwarder/api/solana"
	"github.com/Aran404/Forwarder/api/types"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/google/uuid"
)

func (c *Client) ForwardFunds(ctx context.Context, pr *PaymentCreateResponse) error {
	path := fmt.Sprintf("wal/%v.dat", pr.Address)
	from, err := c.sol.FromFile(path)
	if err != nil {
		return err
	}

	tx, err := c.sol.SendAllBalance(ctx, from, types.Config.Forwarder.ForwardAddress, false)
	if err != nil {
		return err
	}

	log.Printf("Successfully forwarded funds from %v to %v. Transaction: %v", pr.Address, types.Config.Forwarder.ForwardAddress, tx.String())
	return from.Dispose(ctx)
}

func (c *Client) HandleWebhookCall(ctx context.Context, b *PaymentCreateBody, r *ws.LogResult, pr *PaymentCreateResponse) bool {
	tx, err := c.sol.GetTransaction(ctx, r.Value.Signature.String())
	if err != nil || tx.Meta.Err != nil {
		return false
	}

	value, err := tx.Value()
	if err != nil {
		return false
	}

	amount := big.NewFloat(b.Amount)
	if new(big.Float).Mul(amount, IgnoreIotaTxThreshold).Cmp(value) >= 0 {
		return false
	}

	fvalue, _ := value.Float64()
	response := &WebhookResponse{
		Success:        true,
		ID:             pr.ID,
		DesiredAmount:  b.Amount,
		AmountSent:     fvalue,
		TransactionID:  r.Value.Signature.String(),
		Address:        pr.Address,
		TimeSent:       uint64(time.Now().Unix()),
		PercentOfTotal: (fvalue / float64(b.Amount)) * 100,
	}

	if new(big.Float).Mul(amount, TransactionThreshold).Cmp(value) >= 0 {
		response.Error = types.GetProperError(types.ErrTransactionSlipped)
	}

	SendWebhook(b.CallbackURI, response)
	_ = c.db.Write(ctx, "transactions", response)

	if err := c.ForwardFunds(ctx, pr); err != nil {
		return false
	}

	return true
}

func (c *Client) HandleCreatePayment(w http.ResponseWriter, r *http.Request, b *PaymentCreateBody) {
	response := &PaymentCreateResponse{
		Success: true,
		ID:      uuid.New().String(),
		Amount:  b.Amount,
		Expires: uint64(time.Now().Add(CryptoDeadline).Unix()),
	}

	front := c.sol.CreateWallet()
	response.Address = front.PublicKey.String()

	if _, err := front.WriteTemporary(); err != nil {
		types.GInternalServerError(w)
		return
	}

	qr, err := solana.CreateQR(response.Address, float64(b.Amount))
	if err != nil {
		types.GInternalServerError(w)
		return
	}
	response.QRCode = qr

	go func() {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(CryptoDeadline))
		defer cancel()

		c.sol.ListenForTX(ctx, response.Address, func(v *ws.LogResult) bool {
			if v.Value.Err != nil {
				return false
			}
			return c.HandleWebhookCall(ctx, b, v, response)
		}, rpc.CommitmentConfirmed)
	}()

	SendJSON(w, response)
}
