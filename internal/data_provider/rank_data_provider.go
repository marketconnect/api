package postgressql

import (
	"context"
	"database/sql"
	"fmt"

	pb "mc_api/pkg/api"

	client "github.com/i-b8o/postgresql_client"
)

const (
	addPhraseQuery     = `INSERT INTO public.phrases (content) VALUES ($1) ON CONFLICT (content) DO NOTHING RETURNING id;`
	addUserPhraseQuery = `INSERT INTO public.mc_user_phrase (user_id, phrase_id) VALUES ($1, $2) ON CONFLICT (user_id, phrase_id) DO NOTHING`
	addPhraseRankQuery = `INSERT INTO public.ranks (mp, geo, act, user_id, phrase_id, rank, paid_rank, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_DATE) ON CONFLICT ON CONSTRAINT unique_mp_user_id_phrase_id_created_at DO UPDATE SET rank = EXCLUDED.rank, paid_rank = EXCLUDED.paid_rank WHERE ranks.created_at = CURRENT_DATE`
	// selectUserPhrases          = `SELECT p.content, COALESCE(r.mp, '') AS mp, COALESCE(r.geo, '') AS geo, COALESCE(r.act, '') AS act, COALESCE(r.rank, 0) AS rank, COALESCE(r.paid_rank, 0) AS paid_rank, COALESCE(r.created_at, '1970-01-01'::date) AS created_at FROM public.mc_user_phrase up JOIN public.phrases p ON up.phrase_id = p.id LEFT JOIN public.ranks r ON up.phrase_id = r.phrase_id AND up.user_id = r.user_id WHERE up.user_id = $1`
	selectUserPhrases          = `SELECT p.id, p.content, COALESCE(r.mp, '') AS mp, COALESCE(r.geo, '') AS geo, COALESCE(r.act, '') AS act, COALESCE(r.rank, 0) AS rank, COALESCE(r.paid_rank, 0) AS paid_rank, COALESCE(r.created_at, '1970-01-01'::date) AS created_at FROM public.mc_user_phrase up JOIN public.phrases p ON up.phrase_id = p.id LEFT JOIN public.ranks r ON up.phrase_id = r.phrase_id AND up.user_id = r.user_id WHERE up.user_id = $1 AND COALESCE(r.mp, '') = $2`
	selectOldRanksByUserAndGeo = `SELECT r.id, r.mp, r.geo, r.user_id, r.phrase_id, r.rank, r.paid_rank, r.created_at, r.updated_at, p.content FROM public.ranks r JOIN public.phrases p ON r.phrase_id = p.id WHERE r.updated_at < NOW() - INTERVAL '23 hours' AND r.user_id BETWEEN $1 AND $2 AND r.geo = $3 ORDER BY updated_at ASC`
)

type rankStorage struct {
	client client.PostgreSQLClient
}

func NewPhraseStorage(client client.PostgreSQLClient) *rankStorage {
	return &rankStorage{client: client}
}
func (ps *rankStorage) AddPhrase(ctx context.Context, content string) (uint64, error) {
	row := ps.client.QueryRow(ctx, addPhraseQuery, content)
	var phraseID uint64
	err := row.Scan(&phraseID)
	return phraseID, err
}
func (ps *rankStorage) AddUserPhrase(ctx context.Context, userID, phraseID uint64) error {
	_, err := ps.client.Exec(ctx, addUserPhraseQuery, userID, phraseID)
	if err != nil {
		return fmt.Errorf("failed to add user phrase: %v", err)
	}
	return nil
}

func (ps *rankStorage) AddPhraseRank(ctx context.Context, mp string, geo string, act string, userID uint64, phraseID uint64, rank uint64, paidRank uint64) error {
	_, err := ps.client.Exec(ctx, addPhraseRankQuery, mp, geo, act, userID, phraseID, rank, paidRank)
	if err != nil {
		return fmt.Errorf("failed to add phrase rank: %v", err)
	}
	return nil
}

func (ps *rankStorage) SelectUserPhrases(ctx context.Context, userID uint64, mp string) ([]*pb.KeyPhrase, error) {
	rows, err := ps.client.Query(ctx, selectUserPhrases, userID, mp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*pb.KeyPhrase
	keyphraseMap := make(map[string]*pb.KeyPhrase)
	for rows.Next() {
		var content, mp, geo, act string
		var id, rank, paidRank sql.NullInt64
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &content, &mp, &geo, &act, &rank, &paidRank, &createdAt); err != nil {
			return nil, err
		}
		if keyphrase, exists := keyphraseMap[content]; exists {
			if rank.Valid && paidRank.Valid {
				keyphrase.Ranks = append(keyphrase.Ranks, &pb.Rank{Mp: mp, Geo: geo, Action: act, Date: createdAt.Time.String(), Rank: int32(rank.Int64), PaidRank: int32(paidRank.Int64)})
			}
		} else {
			keyphrase := &pb.KeyPhrase{Phrase: &pb.Phrase{Id: uint64(id.Int64), Text: content}}
			if rank.Valid && paidRank.Valid {
				keyphrase.Ranks = []*pb.Rank{{Mp: mp, Geo: geo, Action: act, Date: createdAt.Time.String(), Rank: int32(rank.Int64), PaidRank: int32(paidRank.Int64)}}
			} else {
				keyphrase.Ranks = []*pb.Rank{}
			}
			keyphraseMap[content] = keyphrase
		}
	}

	for _, keyphrase := range keyphraseMap {
		result = append(result, keyphrase)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (ps *rankStorage) SelectOldRanks(ctx context.Context, startID, endID uint64, geo string) ([]*pb.OldRank, error) {
	rows, err := ps.client.Query(ctx, selectOldRanksByUserAndGeo, startID, endID, geo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*pb.OldRank
	for rows.Next() {
		var userID, phraseID sql.NullInt64
		var mp, content sql.NullString
		if err := rows.Scan(&mp, &userID, &phraseID, &content); err != nil {
			return nil, err
		}
		r := &pb.OldRank{
			UserId:   uint64(userID.Int64),
			Phrase:   content.String,
			Products: []string{mp.String},
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
