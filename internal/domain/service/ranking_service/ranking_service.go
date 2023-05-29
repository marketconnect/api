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
	AddPhraseRank(ctx context.Context, userID, phraseID, rank, paidRank uint64) error
	SelectUserPhrases(ctx context.Context, userID uint64) ([]*pb.KeyPhrase, error)
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
