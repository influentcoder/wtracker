package config

import (
	"os"
	"time"
)

// WhaleDef defines a tracked whale address with its label and chain.
type WhaleDef struct {
	Address string
	Label   string
	Chain   string
}

// Config holds application configuration.
type Config struct {
	Port             string
	Whales           []WhaleDef
	BalanceCacheTTL  time.Duration
	TxCacheTTL       time.Duration
	PriceCacheTTL    time.Duration
}

// Load returns the application config, reading PORT from the environment.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:            port,
		BalanceCacheTTL: 5 * time.Minute,
		TxCacheTTL:      5 * time.Minute,
		PriceCacheTTL:   2 * time.Minute,
		Whales:          defaultWhales,
	}
}

// defaultWhales is the initial set of tracked Bitcoin whale addresses.
var defaultWhales = []WhaleDef{
	{
		Address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
		Label:   "Satoshi - Genesis Block",
		Chain:   "bitcoin",
	},
	{
		Address: "34xp4vRoCGJym3xR7yCVPFHoCNxv4Twseo",
		Label:   "Binance Cold Wallet",
		Chain:   "bitcoin",
	},
	{
		Address: "3LYJfcfHcvFMJKHKENnmqeLQnJHNGkPpLx",
		Label:   "Kraken Exchange",
		Chain:   "bitcoin",
	},
	{
		Address: "385cR5DM96n1HvBDMnLjTpFs4iXyofo3Zy",
		Label:   "Huobi Exchange",
		Chain:   "bitcoin",
	},
	{
		Address: "bc1qa5wkgaew2dkv56kfvj49j0av5nml45x9ek9hz6",
		Label:   "Bitfinex",
		Chain:   "bitcoin",
	},
	{
		Address: "1FeexV6bAHb8ybZjqQMjJrcCrHGW9sb6uF",
		Label:   "Mt. Gox Trustee",
		Chain:   "bitcoin",
	},
}
