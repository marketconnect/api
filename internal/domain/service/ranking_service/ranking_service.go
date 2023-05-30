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

type ProductStorage interface {
	SelectUserProducts(ctx context.Context, userID uint64) ([]uint64, error)
}

type RankingService struct {
	rankingStorage RankingStorage
	productStorage ProductStorage
	logging        logging.Logger
	pb.UnimplementedRankServiceServer
}

func NewRankingService(rankingStorage RankingStorage, productStorage ProductStorage, logging logging.Logger) *RankingService {
	return &RankingService{
		rankingStorage: rankingStorage,
		productStorage: productStorage,
		logging:        logging,
	}
}

func (s *RankingService) AddPhrases(ctx context.Context, req *pb.AddPhrasesReq) (*pb.Empty, error) {
	userID, err := mc_jwt.GetIdFromToken(req.Token)
	if err != nil {
		return nil, err
	}
	fmt.Printf("User ID: %d", userID)
	for _, phrase := range req.Phrases {
		phraseID, err := s.rankingStorage.AddPhrase(ctx, phrase.Text)
		if err != nil {
			s.logging.Error(err)
			continue
		}
		err = s.rankingStorage.AddUserPhrase(ctx, userID, phraseID)
		if err != nil {
			s.logging.Error(err)
		}
	}

	return &pb.Empty{}, nil
}

func (s *RankingService) Ranking(ctx context.Context, req *pb.RankingReq) (*pb.RankingResp, error) {
	userID, err := mc_jwt.GetIdFromToken(req.Token)
	if err != nil {
		return nil, err
	}

	keyPhrases, err := s.rankingStorage.SelectUserPhrases(ctx, userID, req.Mp)
	if err != nil {
		return nil, err
	}

	return &pb.RankingResp{
		KeyPhrases: keyPhrases,
	}, nil
}

func (s *RankingService) AddRank(ctx context.Context, req *pb.AddRankReq) (*pb.Empty, error) {
	err := s.rankingStorage.AddPhraseRank(ctx, req.UserId, req.PhraseId, uint64(req.Rank), uint64(req.PaidRank), req.Mp)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}
