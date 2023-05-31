package rank_service

import (
	"context"
	"fmt"
	mc_jwt "mc_api/internal/domain/jwt"
	pb "mc_api/pkg/api"

	"github.com/i-b8o/logging"
)

type RankStorage interface {
	AddPhrase(ctx context.Context, content string) (uint64, error)
	AddUserPhrase(ctx context.Context, userID, phraseID uint64) error
	AddPhraseRank(ctx context.Context, userID, phraseID, rank, paidRank uint64, mp string) error
	SelectUserPhrases(ctx context.Context, userID uint64, mp string) ([]*pb.KeyPhrase, error)
	SelectOldRanks(ctx context.Context, startID, endID uint64) ([]*pb.OldRank, error)
}

type RankService struct {
	rankStorage RankStorage

	logging logging.Logger
	pb.UnimplementedRankServiceServer
}

func NewRankingService(rankStorage RankStorage, logging logging.Logger) *RankService {
	return &RankService{
		rankStorage: rankStorage,
		logging:     logging,
	}
}

func (s *RankService) AddPhrases(ctx context.Context, req *pb.AddPhrasesReq) (*pb.Empty, error) {
	userID, err := mc_jwt.GetIdFromToken(req.Token)
	if err != nil {
		return nil, err
	}
	fmt.Printf("User ID: %d", userID)
	for _, phrase := range req.Phrases {
		phraseID, err := s.rankStorage.AddPhrase(ctx, phrase.Text)
		if err != nil {
			s.logging.Error(err)
			continue
		}
		err = s.rankStorage.AddUserPhrase(ctx, userID, phraseID)
		if err != nil {
			s.logging.Error(err)
		}
	}

	return &pb.Empty{}, nil
}

func (s *RankService) Ranking(ctx context.Context, req *pb.RankingReq) (*pb.RankingResp, error) {
	userID, err := mc_jwt.GetIdFromToken(req.Token)
	if err != nil {
		return nil, err
	}

	keyPhrases, err := s.rankStorage.SelectUserPhrases(ctx, userID, req.Mp)
	if err != nil {
		return nil, err
	}

	return &pb.RankingResp{
		KeyPhrases: keyPhrases,
	}, nil
}

func (s *RankService) AddRank(ctx context.Context, req *pb.AddRankReq) (*pb.Empty, error) {
	err := s.rankStorage.AddPhraseRank(ctx, req.UserId, req.PhraseId, uint64(req.Rank), uint64(req.PaidRank), req.Mp)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (s *RankService) OldRanks(ctx context.Context, req *pb.OldRanksReq) (*pb.OldRanksResp, error) {
	oldRanks, err := s.rankStorage.SelectOldRanks(ctx, uint64(req.From), uint64(req.To))
	if err != nil {
		s.logging.Error(err)
		return nil, err
	}

	return &pb.OldRanksResp{OldRanks: oldRanks}, nil
}
