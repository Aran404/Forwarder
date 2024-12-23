package solana

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"

	"github.com/gagliardetto/solana-go"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"golang.org/x/exp/constraints"
)

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

type Number interface {
	constraints.Integer | constraints.Float
}

func ConvertLamportToSol[T Number](lamports T) *big.Float {
	totalLamports := new(big.Float).SetUint64(uint64(lamports))
	lmps := new(big.Float).Quo(totalLamports, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))
	return lmps
}

func createQR(data string) (string, error) {
	qrc, err := qrcode.New(data)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	wr := nopCloser{Writer: buf}
	if err = qrc.Save(standard.NewWithWriter(wr, standard.WithQRWidth(40))); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// CreateSoloQR is a function that creates a QR code for a given address and returns the base64 representation
func CreateSoloQR(address string) (string, error) {
	raw := fmt.Sprintf("solana:%s", address)
	return createQR(raw)
}

// CreateQR is a function that creates a QR code for a given address and amount
func CreateQR(address string, amount float64) (string, error) {
	raw := fmt.Sprintf("solana:%s?amount=%.9f", address, amount) // 18446744073.709551615
	return createQR(raw)
}
