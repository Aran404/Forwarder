package solana

import (
	"context"
	"log"
	"time"

	"github.com/Aran404/Forwarder/api/types"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"golang.org/x/time/rate"
)

type TransactionBundle struct {
	From   *walletPair
	To     string  // address
	Amount float64 // in SOL
}

type ledgerResult struct {
	*rpc.GetParsedTransactionResult
}

type simulationResult struct {
	*rpc.SimulateTransactionResult
}

type walletPair struct {
	PrivateKey solana.PrivateKey
	PublicKey  solana.PublicKey
}

type jsonWallet struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type Client struct {
	httpNode string
	wsNode   string

	rpc *rpc.Client
	ws  *ws.Client
}

func NewClient(ctx context.Context) *Client {
	rpc := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		types.Env.SOLANA_NET_HTTP,
		rate.Every(time.Duration(types.Config.RatelimitEvery)*time.Second),
		types.Config.RatelimitReset,
	))
	wsClient, err := ws.Connect(ctx, types.Env.SOLANA_NET_WS)
	if err != nil {
		log.Fatal(err)
	}
	return &Client{
		rpc:      rpc,
		ws:       wsClient,
		httpNode: types.Env.SOLANA_NET_HTTP,
		wsNode:   types.Env.SOLANA_NET_WS,
	}
}
