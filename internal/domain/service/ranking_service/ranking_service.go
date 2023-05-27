package ranking_service

import (
	pb "mc_api/pkg/api"

	"github.com/i-b8o/logging"
)

type RankingStorage interface {
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
