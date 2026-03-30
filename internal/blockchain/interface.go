package blockchain

import (
	"context"
	"time"
)

// Client is the chain-agnostic interface for querying blockchain data.
// Implement this interface to add support for additional blockchains.
type Client interface {
	GetAddressInfo(ctx context.Context, address string) (*AddressInfo, error)
	GetRecentTransactions(ctx context.Context, address string, limit int) ([]Transaction, error)
}

// AddressInfo holds summary data for a blockchain address.
type AddressInfo struct {
	Address    string
	BalanceSat int64 // confirmed balance in satoshis (or smallest unit)
	TxCount    int
	Chain      string
}

// Transaction represents a single on-chain transaction relative to a watched address.
type Transaction struct {
	TxID      string
	Timestamp time.Time
	AmountSat int64  // positive = received, negative = sent (in satoshis)
	FeeSat    int64
	Confirmed bool
	BlockHeight int64
}
