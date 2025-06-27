package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/domain"
)

type DataManager struct {
	pool *pgxpool.Pool
}

func NewDataManager(db *pgxpool.Pool) *DataManager {
	return &DataManager{db}
}

func (p *DataManager) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (last_name, first_name, middle_name, user_group)
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`
	err := p.pool.QueryRow(ctx, query,
		user.LastName,
		user.FirstName,
		user.MiddleName,
		user.Group).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (p *DataManager) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET last_name = $1, first_name = $2, middle_name = $3, user_group = $4
		WHERE id = $5
	`

	_, err := p.pool.Exec(ctx, query,
		user.LastName,
		user.FirstName,
		user.MiddleName,
		user.Group,
		user.ID)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (p *DataManager) GetUserById(ctx context.Context, id int) (*domain.User, error) {
	query := `
		SELECT id, last_name, first_name, middle_name, user_group
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := p.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.LastName,
		&user.FirstName,
		&user.MiddleName,
		&user.Group)

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (p *DataManager) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := p.pool.Exec(ctx, query, id)

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (p *DataManager) ListUsers(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, last_name, first_name, middle_name, user_group
		FROM users
		ORDER BY id
	`

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.LastName,
			&user.FirstName,
			&user.MiddleName,
			&user.Group,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}
