package domain

import (
	"github.com/govalues/money"
)

type System struct {
	Num_CPU    int                `json:"num_cpu"`
	CPU_usage  map[string]float64 `json:"cpu_usage"`
	RAM        int64              `json:"ram"`
	RAM_used   int64              `json:"ram_used"`
	DISC       float64            `json:"disc"`
	DISC_used  float64            `json:"disc_used"`
	GOMAXPROCS int                `json:"gomaxprocs"`
}

// Стандартный объект товара
type Item struct {
	SKU      uint64         `json:"sku"`
	Name     string         `json:"name"`
	Price    money.Amount   `json:"price"`
	Currency money.Currency `json:"currency"`
	Quantity int            `json:"quantity"`
}

// Объект товара для возврата ручек
type ItemDTO struct {
	SKU      uint64        `json:"sku"`
	Name     string        `json:"name"`
	Price    MoneySplitted `json:"price"`
	Currency string        `json:"currency"`
	Quantity int           `json:"quantity"`
}

// Объект money для маршалинга в json
type MoneySplitted struct {
	Whole    int64 `json:"whole"`
	Fracture int64 `json:"fracture"`
}
