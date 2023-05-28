package ranking_service

import (
	"context"
	pb "mc_api/pkg/api"

	"github.com/i-b8o/logging"
)

type RankingStorage interface {
	AddKeyPhrases(ctx context.Context, keyPhrases []string) error
	GetKeyPhrases(ctx context.Context, keyPhrases []string) error
}

type RankingService struct {
	storage RankingStorage
	logging logging.Logger
	pb.UnimplementedAPIServer
}

func NewRankingService(storage RankingStorage, logging logging.Logger) *RankingService {
	return &RankingService{
		storage: storage,
		logging: logging,
	}
}
