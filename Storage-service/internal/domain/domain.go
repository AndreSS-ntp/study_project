package domain

import (
	"fmt"
	"github.com/govalues/decimal"
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
	SKU      uint64       `json:"sku"`
	Name     string       `json:"name"`
	Price    money.Amount `json:"price"`
	Quantity int          `json:"quantity"`
}

// Паттерн DTO
type ItemDTO struct {
	SKU      uint64          `json:"sku"`
	Name     string          `json:"name"`
	Price    decimal.Decimal `json:"price"`
	Currency money.Currency  `json:"currency"`
	Quantity int             `json:"quantity"`
}

func ToItemDTO(item *Item) *ItemDTO {
	return &ItemDTO{
		SKU:      item.SKU,
		Name:     item.Name,
		Price:    item.Price.Decimal(),
		Currency: item.Price.Curr(),
		Quantity: item.Quantity,
	}
}

func ToItem(itemDTO *ItemDTO) (*Item, error) {
	price, err := money.NewAmountFromDecimal(itemDTO.Currency, itemDTO.Price)
	if err != nil {
		return nil, fmt.Errorf("parse money.Amount from decimal: %w", err)
	}
	return &Item{
		SKU:      itemDTO.SKU,
		Name:     itemDTO.Name,
		Price:    price,
		Quantity: itemDTO.Quantity,
	}, nil
}
