package db

import (
	"database/sql"
	"errors"

	"github.com/realestate-trust/monorepo/internal/core"
)

type SQLTokenizationRepository struct {
	db *sql.DB
}

func NewSQLTokenizationRepository(db *sql.DB) *SQLTokenizationRepository {
	return &SQLTokenizationRepository{db: db}
}

func (r *SQLTokenizationRepository) CreatePool(propertyID string, totalTokens int64, price float64) (*core.FractionalPool, error) {
	id := "pool-" + propertyID
	query := `INSERT INTO fractional_pools (id, property_id, total_tokens, tokens_sold, token_price)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, property_id, total_tokens, tokens_sold, token_price`

	pool := &core.FractionalPool{}
	err := r.db.QueryRow(query, id, propertyID, totalTokens, 0, price).Scan(
		&pool.ID,
		&pool.PropertyID,
		&pool.TotalTokens,
		&pool.TokensSold,
		&pool.TokenPrice,
	)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func (r *SQLTokenizationRepository) GetPool(id string) (*core.FractionalPool, error) {
	query := `SELECT id, property_id, total_tokens, tokens_sold, token_price FROM fractional_pools WHERE id = $1`
	pool := &core.FractionalPool{}
	err := r.db.QueryRow(query, id).Scan(
		&pool.ID,
		&pool.PropertyID,
		&pool.TotalTokens,
		&pool.TokensSold,
		&pool.TokenPrice,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("fractional pool not found")
		}
		return nil, err
	}
	return pool, nil
}

func (r *SQLTokenizationRepository) BuyTokens(poolID, investorID string, count int64) (*core.FractionalHolding, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Get pool and lock row
	queryPool := `SELECT id, property_id, total_tokens, tokens_sold, token_price FROM fractional_pools WHERE id = $1 FOR UPDATE`
	pool := &core.FractionalPool{}
	err = tx.QueryRow(queryPool, poolID).Scan(
		&pool.ID,
		&pool.PropertyID,
		&pool.TotalTokens,
		&pool.TokensSold,
		&pool.TokenPrice,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("fractional pool not found")
		}
		return nil, err
	}

	if err := pool.ValidateTokenPurchase(count); err != nil {
		return nil, err
	}

	// Update pool TokensSold
	updatePool := `UPDATE fractional_pools SET tokens_sold = tokens_sold + $1 WHERE id = $2`
	_, err = tx.Exec(updatePool, count, poolID)
	if err != nil {
		return nil, err
	}

	holdingID := "hold-" + poolID + "-" + investorID
	// Upsert holding
	upsertHolding := `INSERT INTO fractional_holdings (id, pool_id, investor_id, token_count)
	                  VALUES ($1, $2, $3, $4)
	                  ON CONFLICT (id) DO UPDATE SET token_count = fractional_holdings.token_count + EXCLUDED.token_count
	                  RETURNING id, pool_id, investor_id, token_count`

	holding := &core.FractionalHolding{}
	err = tx.QueryRow(upsertHolding, holdingID, poolID, investorID, count).Scan(
		&holding.ID,
		&holding.PoolID,
		&holding.InvestorID,
		&holding.TokenCount,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return holding, nil
}

func (r *SQLTokenizationRepository) ListPools() ([]*core.FractionalPool, error) {
	query := `SELECT id, property_id, total_tokens, tokens_sold, token_price FROM fractional_pools`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []*core.FractionalPool
	for rows.Next() {
		pool := &core.FractionalPool{}
		err := rows.Scan(
			&pool.ID,
			&pool.PropertyID,
			&pool.TotalTokens,
			&pool.TokensSold,
			&pool.TokenPrice,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, pool)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return []*core.FractionalPool{}, nil
	}
	return result, nil
}
