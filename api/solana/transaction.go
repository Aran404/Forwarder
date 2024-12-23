package solana

import (
	"context"
	"fmt"

	"github.com/Aran404/Forwarder/api/types"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

// CreateMTransactions creates multiple atomic transactions that are bundled together
// If one of the transactions fails, the entire bundle fails
func (c Client) CreateMTransactions(ctx context.Context, tb []*TransactionBundle, payer *walletPair, simulate bool) (*solana.Signature, error) {
	if len(tb) == 0 {
		return nil, nil
	}
	if payer == nil {
		payer = tb[0].From
	}

	tx, err := c.buildTransactions(ctx, tb, payer)
	if err != nil {
		return nil, err
	}

	if simulate {
		sim, err := c.SimulateTransaction(ctx, tx)
		if err != nil {
			return nil, err
		}
		if sim.Overboard() {
			return nil, types.ErrTransactionOverboard
		}
	}

	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentConfirmed,
	}
	sig, err := c.rpc.SendTransactionWithOpts(ctx, tx, opts)
	return &sig, err
}

// CreateTransaction creates a singly atomic transaction
func (c Client) CreateTransaction(ctx context.Context, from *walletPair, to string, amount float64, simulate bool) (*solana.Signature, error) {
	return c.CreateMTransactions(
		ctx,
		[]*TransactionBundle{
			{From: from, To: to, Amount: amount},
		},
		nil,
		simulate,
	)
}

// SendAllBalance sends the entire balance of a wallet
func (c Client) SendAllBalance(ctx context.Context, from *walletPair, to string, simulate bool) (*solana.Signature, error) {
	bal, err := c.WalletBalance(ctx, from.PublicKey.String())
	if err != nil {
		return nil, err
	}

	balance, _ := bal.Float64()
	tb := []*TransactionBundle{{From: from, To: to, Amount: balance}}
	tx, err := c.buildTransactions(ctx, tb, from)
	if err != nil {
		return nil, err
	}

	fee, err := c.rpc.GetFeeForMessage(ctx, tx.Message.ToBase64(), rpc.CommitmentConfirmed)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction fee: %v", err)
	}

	if fee == nil || fee.Value == nil || *fee.Value <= 0 {
		return nil, fmt.Errorf("failed to get transaction fee")
	}

	totalLamports := uint64(balance * float64(solana.LAMPORTS_PER_SOL))
	if totalLamports <= *fee.Value {
		return nil, fmt.Errorf("insufficient funds to cover transaction fee: balance=%d, fee=%d", totalLamports, fee)
	}

	tb[0].Amount = float64(totalLamports-*fee.Value) / float64(solana.LAMPORTS_PER_SOL)
	tx, err = c.buildTransactions(ctx, tb, from)
	if err != nil {
		return nil, err
	}

	if simulate {
		sim, err := c.SimulateTransaction(ctx, tx)
		if err != nil {
			return nil, err
		}
		if sim.Overboard() {
			return nil, types.ErrTransactionOverboard
		}
	}

	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentConfirmed,
	}
	sig, err := c.rpc.SendTransactionWithOpts(ctx, tx, opts)
	return &sig, err
}

// SimulateTransaction simulates a transaction
func (c Client) SimulateTransaction(ctx context.Context, tx *solana.Transaction) (*simulationResult, error) {
	sim, err := c.rpc.SimulateTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	sr := &simulationResult{sim.Value}
	return sr, sr.Error()
}

// SimulationResult returns the simulation result
func (s simulationResult) Failed() bool { return s.Err != nil }

// SimulationError returns the simulation error
func (s simulationResult) Error() error {
	if s.Err == nil {
		return nil
	}
	return fmt.Errorf("Simulation Failed: %v", s.Err)
}

// SimulationOverboard returns true if the simulation went overboard
func (s simulationResult) Overboard() bool {
	return s.Error() != nil || s.UnitsConsumed != nil && *s.UnitsConsumed > 1400000
}

func (c Client) mapWallets(tb []*TransactionBundle) map[solana.PublicKey]*solana.PrivateKey {
	m := make(map[solana.PublicKey]*solana.PrivateKey)
	for _, t := range tb {
		m[t.From.PublicKey] = &t.From.PrivateKey
	}
	return m
}

func (c Client) buildTransactions(ctx context.Context, tb []*TransactionBundle, payer *walletPair) (*solana.Transaction, error) {
	recent, err := c.rpc.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, err
	}

	var (
		instructions []solana.Instruction
		consumed     uint64
	)
	for _, t := range tb {
		singly := system.NewTransferInstruction(
			uint64(t.Amount*float64(solana.LAMPORTS_PER_SOL)),
			t.From.PublicKey,
			solana.MustPublicKeyFromBase58(t.To),
		).Build()

		b, err := singly.Data()
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, singly)
		consumed += uint64(len(b))
	}

	if consumed > 1232 || len(instructions) > 30 {
		return nil, types.ErrTransactionOverboard
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(payer.PublicKey),
	)

	if err != nil {
		return nil, err
	}

	mappedWallets := c.mapWallets(tb)
	if len(mappedWallets) > 18 {
		return nil, types.ErrTransactionOverboard
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if pk, ok := mappedWallets[key]; ok {
				return pk
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
