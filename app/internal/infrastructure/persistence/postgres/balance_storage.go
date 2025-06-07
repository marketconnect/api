package postgres

import (
	"context"

	"github.com/marketconnect/db_client/postgresql"
)

// BalanceStorage provides methods to manage API key balances in PostgreSQL.
type BalanceStorage struct {
	client postgresql.PostgreSQLClient
}

// NewBalanceStorage creates a new BalanceStorage instance.
func NewBalanceStorage(client postgresql.PostgreSQLClient) *BalanceStorage {
	return &BalanceStorage{client: client}
}

// GetBalance retrieves balance for the given apiKey.
func (s *BalanceStorage) GetBalance(ctx context.Context, apiKey string) (int, error) {
	const query = "SELECT balance FROM api_key_balances WHERE api_key = $1"
	row := s.client.QueryRow(ctx, query, apiKey)
	var balance int
	if err := row.Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}

// SetBalance inserts or updates balance for the given apiKey.
func (s *BalanceStorage) SetBalance(ctx context.Context, apiKey string, balance int) error {
	const query = `INSERT INTO api_key_balances (api_key, balance)
                    VALUES ($1, $2)
                    ON CONFLICT (api_key) DO UPDATE SET balance = EXCLUDED.balance`
	_, err := s.client.Exec(ctx, query, apiKey, balance)
	return err
}

// GetTokenCost returns token cost for the specified token type.
func (s *BalanceStorage) GetTokenCost(ctx context.Context, tokenType string) (int, error) {
	const query = "SELECT cost FROM token_costs WHERE token_type = $1"
	row := s.client.QueryRow(ctx, query, tokenType)
	var cost int
	if err := row.Scan(&cost); err != nil {
		return 0, err
	}
	return cost, nil
}
