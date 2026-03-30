package bitcoin

// blockstreamAddress is the Blockstream API response for /api/address/{addr}
type blockstreamAddress struct {
	Address string `json:"address"`
	ChainStats struct {
		FundedTxoSum int64 `json:"funded_txo_sum"`
		SpentTxoSum  int64 `json:"spent_txo_sum"`
		TxCount      int   `json:"tx_count"`
	} `json:"chain_stats"`
	MempoolStats struct {
		FundedTxoSum int64 `json:"funded_txo_sum"`
		SpentTxoSum  int64 `json:"spent_txo_sum"`
		TxCount      int   `json:"tx_count"`
	} `json:"mempool_stats"`
}

// blockstreamTx is the Blockstream API response item from /api/address/{addr}/txs
type blockstreamTx struct {
	TxID string `json:"txid"`
	Fee  int64  `json:"fee"`
	Status struct {
		Confirmed   bool  `json:"confirmed"`
		BlockHeight int64 `json:"block_height"`
		BlockTime   int64 `json:"block_time"`
	} `json:"status"`
	Vin []struct {
		Prevout struct {
			ScriptpubkeyAddress string `json:"scriptpubkey_address"`
			Value               int64  `json:"value"`
		} `json:"prevout"`
	} `json:"vin"`
	Vout []struct {
		ScriptpubkeyAddress string `json:"scriptpubkey_address"`
		Value               int64  `json:"value"`
	} `json:"vout"`
}
