package domain

import "github.com/govalues/money"

type System struct {
	Num_CPU    int                `json:"num_cpu"`
	CPU_usage  map[string]float64 `json:"cpu_usage"`
	RAM        int64              `json:"ram"`
	RAM_used   int64              `json:"ram_used"`
	DISC       float64            `json:"disc"`
	DISC_used  float64            `json:"disc_used"`
	GOMAXPROCS int                `json:"gomaxprocs"`
}

type Item struct {
	SKU      uint64       `json:"sku"`
	Name     string       `json:"name"`
	Price    money.Amount `json:"price"`
	Quantity int          `json:"quantity"`
}
