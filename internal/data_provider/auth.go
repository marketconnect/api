package postgressql

import (
	"context"
	"errors"
	"fmt"

	client "github.com/i-b8o/postgresql_client"
)

const (
	// registerUser = `INSERT INTO chapter ("name", "num", "order_num","doc_id", "title", "description", "keywords") VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (doc_id, order_num) DO UPDATE SET	name = excluded.name, num = excluded.num,title = excluded.title, description = excluded.description, keywords = excluded.keywords RETURNING "id";`
	registerUserQuery = `INSERT INTO public.mc_user (email, password) VALUES ($1, $2) RETURNING id`
	loginUserQuery    = `SELECT id FROM public.mc_user WHERE email = $1 RETURNING id`
	updatePswdQuery   = `UPDATE public.mc_user SET password = $1 WHERE id = $2`
	deleteUserQuery   = `DELETE FROM public.mc_user WHERE id = $1`
)

type authStorage struct {
	client client.PostgreSQLClient
}

func NewAuthStorage(client client.PostgreSQLClient) *authStorage {
	return &authStorage{client: client}
}

func (as *authStorage) RegisterUser(ctx context.Context, email, password string) (uint64, error) {

	row := as.client.QueryRow(ctx, registerUserQuery, email, password)
	var userID uint64
	err := row.Scan(&userID)

	return userID, err
}

func (as *authStorage) LoginUser(ctx context.Context, email, password string) (uint64, error) {
	var userID uint64
	err := as.client.QueryRow(ctx, loginUserQuery, email).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve user: %v", err)
	}
	// check if password matches
	var queriedPassword string
	err = as.client.QueryRow(ctx, "SELECT password FROM public.user WHERE email=$1", email).Scan(&queriedPassword)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve user password: %v", err)
	}
	if queriedPassword != password {
		return 0, errors.New("incorrect password")
	}
	return userID, nil
}

func (as *authStorage) UpdatePswd(ctx context.Context, id uint64, newPassword string) error {
	_, err := as.client.Exec(ctx, updatePswdQuery, newPassword, id)
	if err != nil {
		return fmt.Errorf("failed to update user password: %v", err)
	}
	return nil
}

func (as *authStorage) DeleteUser(ctx context.Context, id uint64) error {
	_, err := as.client.Exec(ctx, deleteUserQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}
