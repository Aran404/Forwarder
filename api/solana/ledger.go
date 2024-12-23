package solana

import (
	"context"
	"encoding/json"
	"math"
	"math/big"

	"github.com/Aran404/Forwarder/api/types"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// ListenForEvent listens for a solana event
func (c Client) ListenForTX(ctx context.Context, address string, callback func(*ws.LogResult) bool, event ...rpc.CommitmentType) error {
	rpcEvent := rpc.CommitmentType("")
	if len(event) > 0 {
		rpcEvent = event[0]
	}

	signature := solana.MustPublicKeyFromBase58(address)
	sub, err := c.ws.LogsSubscribeMentions(signature, rpcEvent)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		v, err := sub.Recv(ctx)
		if err != nil {
			return err
		}

		if callback(v) {
			return nil
		}
	}
}

// GetConfirmations returns the number of confirmations
func (c Client) GetConfirmations(ctx context.Context, txID string) (int16, error) {
	signature := solana.MustSignatureFromBase58(txID)
	statuses, err := c.rpc.GetSignatureStatuses(ctx, true, signature)
	if err != nil {
		return -1, err
	}

	if len(statuses.Value) > 0 && statuses.Value[0] != nil {
		status := statuses.Value[0]
		if status.ConfirmationStatus == "finalized" || status.ConfirmationStatus == "confirmed" {
			return math.MaxInt16, nil
		}

		if status.Confirmations == nil {
			return 0, types.ErrInvalidStatus
		}

		return int16(*status.Confirmations), nil
	}

	return 0, nil
}

// GetTransaction returns the raw transaction
func (c Client) GetTransaction(ctx context.Context, txID string) (*ledgerResult, error) {
	version := uint64(0)
	tx, err := c.rpc.GetParsedTransaction(
		ctx,
		solana.MustSignatureFromBase58(txID),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: &version,
		},
	)
	if err != nil {
		return nil, err
	}
	return &ledgerResult{tx}, nil
}

// GetTransactionValue returns the value of the transaction in SOL
func (tx ledgerResult) Value() (*big.Float, error) {
	for _, k := range tx.Transaction.Message.Instructions {
		if k.Parsed == nil {
			continue
		}

		raw, err := k.Parsed.MarshalJSON()
		if err != nil {
			return nil, err
		}

		var parsed *rpc.InstructionInfo
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return nil, err
		}

		if parsed.InstructionType == "transfer" {
			// Lamports can overflow float64
			if lamports, ok := parsed.Info["lamports"].(float64); ok {
				return ConvertLamportToSol(lamports), nil
			}
		}
	}

	return nil, nil
}

// From returns the address of the sender
func (tx ledgerResult) From() string {
	return tx.Transaction.Message.AccountKeys[0].PublicKey.String()
}

// To returns the address of the receiver
func (tx ledgerResult) To() string {
	return tx.Transaction.Message.AccountKeys[1].PublicKey.String()
}

// Fee returns the transaction fee
func (tx ledgerResult) Fee() float64 {
	return float64(tx.Meta.Fee) / float64(solana.LAMPORTS_PER_SOL)
}
