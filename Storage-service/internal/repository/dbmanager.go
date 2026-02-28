package repository

import (
	"context"
	"errors"
	"fmt"
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

func (p *DataManager) CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	query := `
		INSERT INTO items (sku, name, price, currency, quantity)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sku) DO NOTHING
		RETURNING sku, name, price, currency, quantity
	`

	skuStr := strconv.FormatUint(item.SKU, 10)
	priceStr := item.Price.Decimal().String()

	row := p.pool.QueryRow(ctx, query,
		skuStr,
		item.Name,
		priceStr,
		item.Price.Curr().String(),
		item.Quantity,
	)

	var (
		dbSKU      string
		dbName     string
		dbPrice    string
		dbCurrency string
		dbQuantity int
	)

	err := row.Scan(&dbSKU, &dbName, &dbPrice, &dbCurrency, &dbQuantity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAlreadyExists
		}
		return nil, fmt.Errorf("scan created item: %w", err)
	}

	parsedSKU, err := strconv.ParseUint(dbSKU, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse sku from db: %w", err)
	}

	amount, err := money.ParseAmount(dbCurrency, dbPrice)
	if err != nil {
		return nil, fmt.Errorf("parse price as money: %w", err)
	}

	return &domain.Item{
		SKU:      parsedSKU,
		Name:     dbName,
		Price:    amount,
		Quantity: dbQuantity,
	}, nil
}

func (p *DataManager) UpdateProduct(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	query := `
		UPDATE items
		SET name = $1, price = $2, currency = $3, quantity = $4
		WHERE sku = $5
		RETURNING sku, name, price, currency, quantity
	`

	skuStr := strconv.FormatUint(item.SKU, 10)
	priceStr := item.Price.Decimal().String()

	var (
		dbSKU      string
		dbName     string
		dbPrice    string
		dbCurrency string
		dbQuantity int
	)

	row := p.pool.QueryRow(ctx, query,
		item.Name,
		priceStr,
		item.Quantity,
		skuStr)

	err := row.Scan(&dbSKU, &dbName, &dbPrice, &dbCurrency, &dbQuantity)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("item not found")
		}
		return nil, fmt.Errorf("scan updated item: %w", err)
	}

	parsedSKU, err := strconv.ParseUint(dbSKU, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse sku from db: %w", err)
	}

	amount, err := money.ParseAmount(dbCurrency, dbPrice)
	if err != nil {
		return nil, fmt.Errorf("parse price as money: %w", err)
	}

	return &domain.Item{
		SKU:      parsedSKU,
		Name:     dbName,
		Price:    amount,
		Quantity: dbQuantity,
	}, nil
}

func (p *DataManager) GetItemBySKU(ctx context.Context, sku uint64) (*domain.Item, error) {
	query := `
		SELECT sku, name, price, currency, quantity
		FROM items
		WHERE sku = $1
	`

	skuStr := strconv.FormatUint(sku, 10)

	var (
		dbSKU      string
		dbName     string
		dbPrice    string
		dbCurrency string
		dbQuantity int
	)
	row := p.pool.QueryRow(ctx, query, skuStr)
	err := row.Scan(
		&dbSKU,
		&dbName,
		&dbPrice,
		&dbCurrency,
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

	amount, err := money.ParseAmount(dbCurrency, dbPrice)
	if err != nil {
		return nil, fmt.Errorf("parse price as money: %w", err)
	}

	return &domain.Item{
		SKU:      parsedSKU,
		Name:     dbName,
		Price:    amount,
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

func (p *DataManager) ListItems(ctx context.Context, limit, offset int) ([]*domain.Item, error) {
	query := `
		SELECT sku, name, price, currency, quantity
		FROM items
		ORDER BY sku
		LIMIT $1 OFFSET $2
	`

	rows, err := p.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	var items []*domain.Item

	for rows.Next() {
		var (
			dbSKU      string
			dbName     string
			dbPrice    string
			dbCurrency string
			dbQuantity int
		)

		if err := rows.Scan(&dbSKU, &dbName, &dbPrice, &dbCurrency, &dbQuantity); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		parsedSKU, err := strconv.ParseUint(dbSKU, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse sku from db: %w", err)
		}

		amount, err := money.ParseAmount(dbCurrency, dbPrice)
		if err != nil {
			return nil, fmt.Errorf("parse price as money: %w", err)
		}

		items = append(items, &domain.Item{
			SKU:      parsedSKU,
			Name:     dbName,
			Price:    amount,
			Quantity: dbQuantity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}
