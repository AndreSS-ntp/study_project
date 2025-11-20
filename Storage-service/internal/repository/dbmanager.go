package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/govalues/decimal"
	"github.com/govalues/money"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/unwisecode/over-the-horison-andress/Storage-service/internal/domain"
	"strconv"
)

type DataManager struct {
	pool *pgxpool.Pool
}

func NewDataManager(db *pgxpool.Pool) *DataManager {
	return &DataManager{db}
}

func (p *DataManager) CreateItem(ctx context.Context, item *domain.Item) (domain.ItemToSend, error) {
	query := `
		INSERT INTO items (sku, name, price, quantity)
		VALUES ($1, $2, $3, $4)
		RETURNING sku, name, price, quantity
	`

	skuStr := strconv.FormatUint(item.SKU, 10)
	priceStr := item.Price.Decimal().String()

	row := p.pool.QueryRow(ctx, query,
		skuStr,
		item.Name,
		priceStr,
		item.Quantity)

	var (
		dbSKU      string
		dbName     string
		dbPrice    string
		dbQuantity int
	)

	err := row.Scan(&dbSKU, &dbName, &dbPrice, &dbQuantity)
	if err != nil {
		return domain.ItemToSend{}, fmt.Errorf("scan created item: %w", err)
	}

	parsedSKU, err := strconv.ParseUint(dbSKU, 10, 64)
	if err != nil {
		return domain.ItemToSend{}, fmt.Errorf("parse sku from db: %w", err)
	}

	amount, err := money.ParseAmount("RUB", dbPrice)
	if err != nil {
		return domain.ItemToSend{}, fmt.Errorf("parse price as money: %w", err)
	}

	whole, fracture, ok := amount.Int64(4)
	if !ok {
		return domain.ItemToSend{}, fmt.Errorf("parse price as MoneySplitted: %w", err)
	}
	currency := amount.Curr()

	return domain.ItemToSend{
		SKU:  parsedSKU,
		Name: dbName,
		Price: domain.MoneySplitted{
			Whole:    whole,
			Fracture: fracture,
			Currency: currency,
		},
		Quantity: dbQuantity,
	}, nil
}

func (p *DataManager) UpdateProduct(ctx context.Context, item *domain.Item) (domain.ItemToSend, error) {
	query := `
		UPDATE items
		SET name = $1, price = $2, quantity = $3
		WHERE sku = $4
		RETURNING sku, name, price, quantity
	`

	skuStr := strconv.FormatUint(item.SKU, 10)
	priceStr := item.Price.Decimal().String()

	var (
		dbSKU      string
		dbName     string
		dbPrice    string
		dbQuantity int
	)

	row := p.pool.QueryRow(ctx, query,
		item.Name,
		priceStr,
		item.Quantity,
		skuStr)

	err := row.Scan(&dbSKU, &dbName, &dbPrice, &dbQuantity)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ItemToSend{}, fmt.Errorf("item not found")
		}
		return domain.ItemToSend{}, fmt.Errorf("scan updated item: %w", err)
	}

	parsedSKU, err := strconv.ParseUint(dbSKU, 10, 64)
	if err != nil {
		return domain.ItemToSend{}, fmt.Errorf("parse sku from db: %w", err)
	}

	amount, err := money.ParseAmount("RUB", dbPrice)
	if err != nil {
		return domain.ItemToSend{}, fmt.Errorf("parse price as money: %w", err)
	}

	whole, fracture, ok := amount.Int64(4)
	if !ok {
		return domain.ItemToSend{}, fmt.Errorf("parse price as MoneySplitted: %w", err)
	}
	currency := amount.Curr()

	return domain.ItemToSend{
		SKU:  parsedSKU,
		Name: dbName,
		Price: domain.MoneySplitted{
			Whole:    whole,
			Fracture: fracture,
			Currency: currency,
		},
		Quantity: dbQuantity,
	}, nil
}

func (p *DataManager) GetItemBySKU(ctx context.Context, sku uint64) (*domain.ItemToSend, error) {
	query := `
		SELECT sku, name, price, quantity
		FROM items
		WHERE sku = $1
	`

	skuStr := strconv.FormatUint(sku, 10)

	var (
		dbSKU      string
		dbName     string
		dbPrice    string
		dbQuantity int
	)
	row := p.pool.QueryRow(ctx, query, skuStr)
	err := row.Scan(
		&dbSKU,
		&dbName,
		&dbPrice,
		&dbQuantity)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("item not found")
		}
		return nil, fmt.Errorf("scan row: %w", err)
	}

	parsedSKU, err := strconv.ParseUint(dbSKU, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse sku from db: %w", err)
	}

	dec, err := decimal.Parse(dbPrice)
	if err != nil {
		return nil, fmt.Errorf("parse price as decimal: %w", err)
	}

	amount, err := money.NewAmountFromDecimal(money.RUB, dec)
	if err != nil {
		return nil, fmt.Errorf("new money amount: %w", err)
	}

	whole, fracture, ok := amount.Int64(4)
	if !ok {
		return nil, fmt.Errorf("parse price as MoneySplitted: %w", err)
	}
	currency := amount.Curr()

	return &domain.ItemToSend{
		SKU:  parsedSKU,
		Name: dbName,
		Price: domain.MoneySplitted{
			Whole:    whole,
			Fracture: fracture,
			Currency: currency,
		},
		Quantity: dbQuantity,
	}, nil
}

func (p *DataManager) DeleteItem(ctx context.Context, sku uint64) error {
	query := `
		DELETE FROM items
		WHERE sku = $1
	`

	skuStr := strconv.FormatUint(sku, 10)
	cmdTag, err := p.pool.Exec(ctx, query, skuStr)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("item not found")
	}

	return nil
}

func (p *DataManager) ListItems(ctx context.Context, limit, offset int) ([]domain.ItemToSend, error) {
	query := `
		SELECT sku, name, price, quantity
		FROM items
		ORDER BY sku
		LIMIT $1 OFFSET $2
	`

	rows, err := p.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	var items []domain.ItemToSend

	for rows.Next() {
		var (
			dbSKU      string
			dbName     string
			dbPrice    string
			dbQuantity int
		)

		if err := rows.Scan(&dbSKU, &dbName, &dbPrice, &dbQuantity); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		parsedSKU, err := strconv.ParseUint(dbSKU, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse sku from db: %w", err)
		}

		dec, err := decimal.Parse(dbPrice)
		if err != nil {
			return nil, fmt.Errorf("parse price as decimal: %w", err)
		}

		amount, err := money.NewAmountFromDecimal(money.RUB, dec)
		if err != nil {
			return nil, fmt.Errorf("new money amount: %w", err)
		}

		whole, fracture, ok := amount.Int64(4)
		if !ok {
			return nil, fmt.Errorf("parse price as MoneySplitted: %w", err)
		}
		currency := amount.Curr()

		items = append(items, domain.ItemToSend{
			SKU:  parsedSKU,
			Name: dbName,
			Price: domain.MoneySplitted{
				Whole:    whole,
				Fracture: fracture,
				Currency: currency,
			},
			Quantity: dbQuantity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}
