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
	AddPhraseRank(ctx context.Context, mp string, geo string, act string, userID uint64, phraseID uint64, rank uint64, paidRank uint64) error
	SelectUserPhrases(ctx context.Context, userID uint64, mp string) ([]*pb.KeyPhrase, error)
	SelectOldRanks(ctx context.Context, startID, endID uint64, geo string) ([]*pb.OldRank, error)
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

func (s *RankService) Rank(ctx context.Context, req *pb.RankingReq) (*pb.RankingResp, error) {
	userID, err := mc_jwt.GetIdFromToken(req.Token)
	if err != nil {
		return nil, err
	}

	keyPhrases, err := s.rankStorage.SelectUserPhrases(ctx, userID, req.Mp)
	if err != nil {
		return nil, err
	}

	// Filter for MP
	// var resKeyPhrases []*pb.KeyPhrase

	// for _, keyPhrase := range keyPhrases {
	// 	var newRanks []*pb.Rank

	// 	for _, r := range keyPhrase.Ranks {
	// 		fmt.Println(r.Rank)
	// 		fmt.Println(r.Rank < 1)
	// 		if r.Mp != req.Mp || r.Rank < 1 {
	// 			fmt.Println(r)
	// 			continue
	// 		}
	// 		newRanks = append(newRanks, r)
	// 	}
	// 	fmt.Println(newRanks)
	// 	newKeyPhrase := &pb.KeyPhrase{Phrase: keyPhrase.Phrase, Ranks: newRanks}
	// 	resKeyPhrases = append(resKeyPhrases, newKeyPhrase)
	// }

	return &pb.RankingResp{
		KeyPhrases: keyPhrases,
	}, nil
}

func (s *RankService) AddRank(ctx context.Context, req *pb.AddRankReq) (*pb.Empty, error) {
	err := s.rankStorage.AddPhraseRank(ctx, req.Mp, req.Geo, req.Action, req.UserId, req.PhraseId, uint64(req.Rank), uint64(req.PaidRank))
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (s *RankService) OldRanks(ctx context.Context, req *pb.OldRanksReq) (*pb.OldRanksResp, error) {
	oldRanks, err := s.rankStorage.SelectOldRanks(ctx, uint64(req.From), uint64(req.To), req.Geo)
	if err != nil {
		s.logging.Error(err)
		return nil, err
	}

	return &pb.OldRanksResp{OldRanks: oldRanks}, nil
}
