package solana

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// CreateWallet create a new wallet
func (Client) CreateWallet() *walletPair {
	account := solana.NewWallet()
	return &walletPair{
		PublicKey:  account.PublicKey(),
		PrivateKey: account.PrivateKey,
	}
}

// WalletBalance returns the balance of a wallet in SOL
func (c Client) WalletBalance(ctx context.Context, address string) (*big.Float, error) {
	signature := solana.MustPublicKeyFromBase58(address)
	bal, err := c.rpc.GetBalance(ctx, signature, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, err
	}
	return ConvertLamportToSol(bal.Value), nil
}

// RequestAirdrop requests an airdrop used for testing
func (c Client) RequestAirdrop(ctx context.Context, w *walletPair) (*solana.Signature, error) {
	sig, err := c.rpc.RequestAirdrop(ctx, w.PublicKey, solana.LAMPORTS_PER_SOL, rpc.CommitmentConfirmed)
	return &sig, err
}

// Encode encodes the wallet to binary
func (w walletPair) Encode() string {
	return w.PrivateKey.String()
}

// WriteTemporary writes the wallet to a temporary file and returns the path
func (w walletPair) WriteTemporary() (string, error) {
	b := w.Encode()
	tmp := fmt.Sprintf("wal/%v.dat", w.PublicKey.String())
	if err := os.WriteFile(tmp, []byte(b), 0600); err != nil {
		return "", err
	}
	return tmp, nil
}

// Dispose disposes of the wallet
// ! Must only be called when the wallet no longer has balance
func (w walletPair) Dispose(ctx context.Context) error {
	tmp := fmt.Sprintf("wal/%v.dat", w.PublicKey.String())
	if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	w.PrivateKey = nil
	w.PublicKey = solana.PublicKey{}
	return nil
}

// FromFile decodes a wallet from a file
func (c Client) FromFile(path string) (*walletPair, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	priv := solana.MustPrivateKeyFromBase58(string(b))
	if priv == nil {
		return nil, fmt.Errorf("invalid private key")
	}

	// kp := strings.Split(path, "/")
	// publicKey := []byte(strings.TrimSuffix(kp[len(kp)-1], ".json"))

	return &walletPair{
		PublicKey:  priv.PublicKey(),
		PrivateKey: priv,
	}, nil
}
