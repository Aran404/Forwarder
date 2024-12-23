package types

import (
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
)

var (
	Env    EnvVars
	Config ConfigVars
)

func init() {
	envmap, err := godotenv.Read()
	if err != nil {
		log.Fatal(err)
	}

	if err := mapstructure.Decode(envmap, &Env); err != nil {
		log.Fatal(err)
	}

	cfg, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}

	defer cfg.Close()
	if err := json.NewDecoder(cfg).Decode(&Config); err != nil {
		log.Fatal(err)
	}
}

type EnvVars struct {
	SOLANA_NET_HTTP string `json:"SOLANA_NET_HTTP" mapstructure:"SOLANA_NET_HTTP"`
	SOLANA_NET_WS   string `json:"SOLANA_NET_WS" mapstructure:"SOLANA_NET_WS"`
}

type ConfigVars struct {
	RatelimitEvery int `json:"ratelimit_every"`
	RatelimitReset int `json:"ratelimit_reset"`
	Forwarder      struct {
		ForwardAddress       string  `json:"foward_address"`
		MinForward           float64 `json:"min_forward"`
		TransactionThreshold float64 `json:"transaction_threshold"`
	} `json:"forwarder"`
}
