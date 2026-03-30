package bitcoin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/influentcoder/wtracker/internal/blockchain"
)

const baseURL = "https://blockstream.info/api"

// Client implements blockchain.Client for Bitcoin using the Blockstream.info API.
type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) GetAddressInfo(ctx context.Context, address string) (*blockchain.AddressInfo, error) {
	url := fmt.Sprintf("%s/address/%s", baseURL, address)
	var resp blockstreamAddress
	if err := c.get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("blockstream get address %s: %w", address, err)
	}

	balance := resp.ChainStats.FundedTxoSum - resp.ChainStats.SpentTxoSum
	txCount := resp.ChainStats.TxCount + resp.MempoolStats.TxCount

	return &blockchain.AddressInfo{
		Address:    address,
		BalanceSat: balance,
		TxCount:    txCount,
		Chain:      "bitcoin",
	}, nil
}

func (c *Client) GetRecentTransactions(ctx context.Context, address string, limit int) ([]blockchain.Transaction, error) {
	url := fmt.Sprintf("%s/address/%s/txs", baseURL, address)
	var raw []blockstreamTx
	if err := c.get(ctx, url, &raw); err != nil {
		return nil, fmt.Errorf("blockstream get txs %s: %w", address, err)
	}

	if limit > 0 && len(raw) > limit {
		raw = raw[:limit]
	}

	txs := make([]blockchain.Transaction, 0, len(raw))
	for _, r := range raw {
		// Calculate net amount for this address
		var received, sent int64
		for _, vout := range r.Vout {
			if vout.ScriptpubkeyAddress == address {
				received += vout.Value
			}
		}
		for _, vin := range r.Vin {
			if vin.Prevout.ScriptpubkeyAddress == address {
				sent += vin.Prevout.Value
			}
		}
		net := received - sent

		var ts time.Time
		if r.Status.BlockTime > 0 {
			ts = time.Unix(r.Status.BlockTime, 0)
		}

		txs = append(txs, blockchain.Transaction{
			TxID:        r.TxID,
			Timestamp:   ts,
			AmountSat:   net,
			FeeSat:      r.Fee,
			Confirmed:   r.Status.Confirmed,
			BlockHeight: r.Status.BlockHeight,
		})
	}
	return txs, nil
}

func (c *Client) get(ctx context.Context, url string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}
