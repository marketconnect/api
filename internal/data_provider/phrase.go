package postgressql

import (
	"context"
	"database/sql"
	"fmt"

	client "github.com/i-b8o/postgresql_client"
)

const (
	addPhraseQuery     = `INSERT INTO public.phrases (content) VALUES ($1) RETURNING id`
	addUserPhraseQuery = `INSERT INTO public.mc_user_phrase (user_id, phrase_id) VALUES ($1, $2)`
	addPhraseRankQuery = `INSERT INTO public.ranks (user_id, phrase_id, rank, paid_rank) VALUES ($1, $2, $3, $4)`
	selectUserPhrases  = `SELECT p.content, r.rank, r.paid_rank FROM public.mc_user_phrase up JOIN public.phrases p ON up.phrase_id = p.id LEFT JOIN public.ranks r ON up.phrase_id = r.phrase_id AND up.user_id = r.user_id WHERE up.user_id = $1`
)

type phraseStorage struct {
	client client.PostgreSQLClient
}

func NewPhraseStorage(client client.PostgreSQLClient) *phraseStorage {
	return &phraseStorage{client: client}
}
func (ps *phraseStorage) AddPhrase(ctx context.Context, content string) (uint64, error) {
	row := ps.client.QueryRow(ctx, addPhraseQuery, content)
	var phraseID uint64
	err := row.Scan(&phraseID)
	return phraseID, err
}
func (ps *phraseStorage) AddUserPhrase(ctx context.Context, userID, phraseID uint64) error {
	_, err := ps.client.Exec(ctx, addUserPhraseQuery, userID, phraseID)
	if err != nil {
		return fmt.Errorf("failed to add user phrase: %v", err)
	}
	return nil
}
func (ps *phraseStorage) AddPhraseRank(ctx context.Context, userID, phraseID, rank, paidRank uint64) error {
	_, err := ps.client.Exec(ctx, addPhraseRankQuery, userID, phraseID, rank, paidRank)
	if err != nil {
		return fmt.Errorf("failed to add phrase rank: %v", err)
	}
	return nil
}
func (ps *phraseStorage) SelectUserPhrases(ctx context.Context, userID uint64) ([]map[string]interface{}, error) {
	rows, err := ps.client.Query(ctx, selectUserPhrases, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var content string
		var rank, paidRank sql.NullInt64
		if err := rows.Scan(&content, &rank, &paidRank); err != nil {
			return nil, err
		}
		row := map[string]interface{}{"content": content}
		if rank.Valid {
			row["rank"] = rank.Int64
		}
		if paidRank.Valid {
			row["paid_rank"] = paidRank.Int64
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
