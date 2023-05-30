package ranking_service

import (
	"context"
	"fmt"
	mc_jwt "mc_api/internal/domain/jwt"
	pb "mc_api/pkg/api"

	"github.com/i-b8o/logging"
)

type RankingStorage interface {
	AddPhrase(ctx context.Context, content string) (uint64, error)
	AddUserPhrase(ctx context.Context, userID, phraseID uint64) error
	AddPhraseRank(ctx context.Context, userID, phraseID, rank, paidRank uint64, mp string) error
	SelectUserPhrases(ctx context.Context, userID uint64, mp string) ([]*pb.KeyPhrase, error)
}

type RankingService struct {
	storage RankingStorage
	logging logging.Logger
	pb.UnimplementedRankServiceServer
}

func NewRankingService(storage RankingStorage, logging logging.Logger) *RankingService {
	return &RankingService{
		storage: storage,
		logging: logging,
	}
}

func (s *RankingService) AddPhrases(ctx context.Context, req *pb.AddPhrasesReq) (*pb.Empty, error) {
	userID, err := mc_jwt.GetIdFromToken(req.Token)
	if err != nil {
		return nil, err
	}
	fmt.Printf("User ID: %d", userID)
	for _, phrase := range req.Phrases {
		phraseID, err := s.storage.AddPhrase(ctx, phrase.Text)
		if err != nil {
			s.logging.Error(err)
			continue
		}
		err = s.storage.AddUserPhrase(ctx, userID, phraseID)
		if err != nil {
			s.logging.Error(err)
		}
	}

	return &pb.Empty{}, nil
}

func (s *RankingService) Ranking(ctx context.Context, req *pb.TokenMessage) (*pb.RankingResp, error) {
	userID, err := mc_jwt.GetIdFromToken(req.Token)
	if err != nil {
		return nil, err
	}
	keyPhrases, err := s.storage.SelectUserPhrases(ctx, userID, "global")
	if err != nil {
		return nil, err
	}
	// var keyPhrases []*pb.KeyPhrase
	// for _, phrase := range phrases {
	// 	var ranks []*pb.Rank
	// 	for _, rank := range phrase.Ranks {
	// 		ranks = append(ranks, &pb.Rank{
	// 			Date:     rank.Date,
	// 			Rank:     int32(rank.Rank),
	// 			PaidRank: int32(rank.PaidRank),
	// 			Place:    rank.Place,
	// 		})
	// 	}
	// 	keyPhrase := &pb.KeyPhrase{
	// 		Phrase: &pb.Phrase{
	// 			Text: phrase.Phrase.Text,
	// 		},
	// 		Ranks: ranks,
	// 	}
	// 	keyPhrases = append(keyPhrases, keyPhrase)
	// }
	return &pb.RankingResp{
		KeyPhrases: keyPhrases,
	}, nil
}

func (s *RankingService) AddRank(ctx context.Context, req *pb.AddRankReq) (*pb.Empty, error) {
	err := s.storage.AddPhraseRank(ctx, req.UserId, req.PhraseId, req.Rank, req.PaidRank, req.Mp)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}
