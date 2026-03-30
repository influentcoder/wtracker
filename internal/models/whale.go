package models

import "time"

// WhaleCard is the summary view of a tracked whale, returned in list responses.
type WhaleCard struct {
	Address    string    `json:"address"`
	Label      string    `json:"label"`
	Chain      string    `json:"chain"`
	BTCBalance float64   `json:"btc_balance"`  // balance in BTC
	USDValue   float64   `json:"usd_value"`    // estimated USD value
	TxCount    int       `json:"tx_count"`
	LastActive time.Time `json:"last_active"`
	Error      string    `json:"error,omitempty"`
}

// WhaleDetail extends WhaleCard with recent transactions.
type WhaleDetail struct {
	WhaleCard
	RecentTransactions []Transaction `json:"recent_transactions"`
}

// Transaction is the app-level transaction model returned by the API.
type Transaction struct {
	TxID        string    `json:"txid"`
	Timestamp   time.Time `json:"timestamp"`
	BTCAmount   float64   `json:"btc_amount"`   // positive = received, negative = sent
	USDAmount   float64   `json:"usd_amount"`
	FeeBTC      float64   `json:"fee_btc"`
	Confirmed   bool      `json:"confirmed"`
	BlockHeight int64     `json:"block_height"`
	Direction   string    `json:"direction"` // "in" or "out"
}

// PriceResponse is returned by /api/price.
type PriceResponse struct {
	BTCUSD    float64   `json:"btc_usd"`
	UpdatedAt time.Time `json:"updated_at"`
}

const SatoshisPerBTC = 1e8
