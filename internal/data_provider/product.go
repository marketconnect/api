package postgressql

import (
	"context"
	"fmt"

	client "github.com/i-b8o/postgresql_client"
)

const (
	addProductQuery         = `INSERT INTO public.mc_products (product, user_id) VALUES ($1, $2) ON CONFLICT (product, user_id) DO NOTHING`
	selectUserProductsQuery = `SELECT product FROM public.mc_products WHERE user_id = $1`
)

type productStorage struct {
	client client.PostgreSQLClient
}

func NewProductStorage(client client.PostgreSQLClient) *productStorage {
	return &productStorage{client: client}
}

func (ps *productStorage) AddProduct(ctx context.Context, userID uint64, product string) error {
	_, err := ps.client.Exec(ctx, addProductQuery, product, userID)
	if err != nil {
		return fmt.Errorf("failed to add product: %v", err)
	}
	return nil
}

func (ps *productStorage) SelectUserProducts(ctx context.Context, userID uint64) ([]uint64, error) {
	rows, err := ps.client.Query(ctx, selectUserProductsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []uint64
	for rows.Next() {
		var productID uint64
		if err := rows.Scan(&productID); err != nil {
			return nil, err
		}
		result = append(result, productID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
