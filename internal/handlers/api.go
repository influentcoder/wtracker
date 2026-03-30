package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/influentcoder/wtracker/internal/blockchain"
	"github.com/influentcoder/wtracker/internal/cache"
	"github.com/influentcoder/wtracker/internal/config"
	"github.com/influentcoder/wtracker/internal/models"
)

// API holds the dependencies for the HTTP handlers.
type API struct {
	cfg     *config.Config
	clients map[string]blockchain.Client // keyed by chain name
	cache   *cache.Cache
}

// NewAPI creates a new API handler set.
func NewAPI(cfg *config.Config, clients map[string]blockchain.Client, c *cache.Cache) *API {
	return &API{cfg: cfg, clients: clients, cache: c}
}

// ListWhales handles GET /api/whales
func (a *API) ListWhales(w http.ResponseWriter, r *http.Request) {
	btcPrice := a.getBTCPrice(r.Context())

	cards := make([]models.WhaleCard, 0, len(a.cfg.Whales))
	for _, def := range a.cfg.Whales {
		card := a.buildWhaleCard(r.Context(), def, btcPrice)
		cards = append(cards, card)
	}
	writeJSON(w, cards)
}

// GetWhale handles GET /api/whales/{address}
func (a *API) GetWhale(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "address")

	var def *config.WhaleDef
	for _, d := range a.cfg.Whales {
		if d.Address == address {
			d := d
			def = &d
			break
		}
	}
	if def == nil {
		http.Error(w, `{"error":"address not tracked"}`, http.StatusNotFound)
		return
	}

	btcPrice := a.getBTCPrice(r.Context())
	card := a.buildWhaleCard(r.Context(), *def, btcPrice)

	txs := a.getTransactions(r.Context(), *def, btcPrice)

	detail := models.WhaleDetail{
		WhaleCard:          card,
		RecentTransactions: txs,
	}
	writeJSON(w, detail)
}

// GetPrice handles GET /api/price
func (a *API) GetPrice(w http.ResponseWriter, r *http.Request) {
	price := a.getBTCPrice(r.Context())
	writeJSON(w, models.PriceResponse{
		BTCUSD:    price,
		UpdatedAt: time.Now(),
	})
}

func (a *API) buildWhaleCard(ctx context.Context, def config.WhaleDef, btcPrice float64) models.WhaleCard {
	card := models.WhaleCard{
		Address: def.Address,
		Label:   def.Label,
		Chain:   def.Chain,
	}

	cacheKey := fmt.Sprintf("info:%s:%s", def.Chain, def.Address)
	if v, ok := a.cache.Get(cacheKey); ok {
		info := v.(*blockchain.AddressInfo)
		card.BTCBalance = float64(info.BalanceSat) / models.SatoshisPerBTC
		card.USDValue = card.BTCBalance * btcPrice
		card.TxCount = info.TxCount
		return card
	}

	client, ok := a.clients[def.Chain]
	if !ok {
		card.Error = "unsupported chain"
		return card
	}

	info, err := client.GetAddressInfo(ctx, def.Address)
	if err != nil {
		log.Printf("GetAddressInfo %s: %v", def.Address, err)
		card.Error = "failed to fetch data"
		return card
	}

	a.cache.Set(cacheKey, info, a.cfg.BalanceCacheTTL)
	card.BTCBalance = float64(info.BalanceSat) / models.SatoshisPerBTC
	card.USDValue = card.BTCBalance * btcPrice
	card.TxCount = info.TxCount
	return card
}

func (a *API) getTransactions(ctx context.Context, def config.WhaleDef, btcPrice float64) []models.Transaction {
	cacheKey := fmt.Sprintf("txs:%s:%s", def.Chain, def.Address)
	if v, ok := a.cache.Get(cacheKey); ok {
		return v.([]models.Transaction)
	}

	client, ok := a.clients[def.Chain]
	if !ok {
		return nil
	}

	raw, err := client.GetRecentTransactions(ctx, def.Address, 10)
	if err != nil {
		log.Printf("GetRecentTransactions %s: %v", def.Address, err)
		return nil
	}

	txs := make([]models.Transaction, 0, len(raw))
	for _, r := range raw {
		btcAmt := float64(r.AmountSat) / models.SatoshisPerBTC
		dir := "in"
		if r.AmountSat < 0 {
			dir = "out"
		}
		txs = append(txs, models.Transaction{
			TxID:        r.TxID,
			Timestamp:   r.Timestamp,
			BTCAmount:   btcAmt,
			USDAmount:   btcAmt * btcPrice,
			FeeBTC:      float64(r.FeeSat) / models.SatoshisPerBTC,
			Confirmed:   r.Confirmed,
			BlockHeight: r.BlockHeight,
			Direction:   dir,
		})
	}

	a.cache.Set(cacheKey, txs, a.cfg.TxCacheTTL)
	return txs
}

func (a *API) getBTCPrice(ctx context.Context) float64 {
	const cacheKey = "price:btc"
	if v, ok := a.cache.Get(cacheKey); ok {
		return v.(float64)
	}

	price, err := fetchBTCPrice(ctx)
	if err != nil {
		log.Printf("fetchBTCPrice: %v", err)
		return 0
	}

	a.cache.Set(cacheKey, price, a.cfg.PriceCacheTTL)
	return price
}

// fetchBTCPrice fetches the current BTC/USD spot price from the Coinbase public API.
// No authentication or API key required.
func fetchBTCPrice(ctx context.Context) (float64, error) {
	url := "https://api.coinbase.com/v2/prices/BTC-USD/spot"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("coinbase price API status %d", resp.StatusCode)
	}

	var data struct {
		Data struct {
			Amount string `json:"amount"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	var price float64
	if _, err := fmt.Sscanf(data.Data.Amount, "%f", &price); err != nil {
		return 0, fmt.Errorf("parsing price %q: %w", data.Data.Amount, err)
	}
	return price, nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}
