package product_service

import (
	"context"
	mc_jwt "mc_api/internal/domain/jwt"
	pb "mc_api/pkg/api"

	"github.com/i-b8o/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductStorage interface {
	AddProduct(ctx context.Context, userID uint64, productID string) error
}

type ProductService struct {
	storage ProductStorage
	logging logging.Logger
	pb.UnimplementedProductServiceServer
}

func NewProductService(storage ProductStorage, logging logging.Logger) *ProductService {
	return &ProductService{
		storage: storage,
		logging: logging,
	}
}

func (s *ProductService) AddProducts(ctx context.Context, req *pb.AddProductsReq) (*pb.Empty, error) {
	userID, err := mc_jwt.GetIdFromToken(req.Token)
	if err != nil {
		return nil, err
	}

	products := req.Products
	for _, product := range products {
		err := s.storage.AddProduct(ctx, userID, product)
		if err != nil {
			s.logging.Errorf(err.Error())
			return nil, status.Error(codes.Internal, "failed to add product to storage")
		}
	}

	return &pb.Empty{}, nil
}
