package service

import (
	"context"

	pb "mc_api/gen/proto"

	"github.com/i-b8o/logging"
)

type AuthStorage interface {
	RegisterUser(ctx context.Context, email, password string) (uint64, error)
	LoginUser(ctx context.Context, email, password string) (uint64, error)
}

type AuthService struct {
	storage AuthStorage
	logging logging.Logger
	pb.UnimplementedAPIServer
}

func NewAuthService(storage AuthStorage, logging logging.Logger) *AuthService {
	return &AuthService{
		storage: storage,
		logging: logging,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, user *pb.User) (*pb.AuthResponse, error) {
	print("HERERERER")
	email := user.GetEmail()
	pswd := user.GetPassword()
	id, err := s.storage.RegisterUser(ctx, email, pswd)
	if err != nil {
		s.logging.Errorf("Error creating user: %v", err)
		return nil, err
	}
	return &pb.AuthResponse{Id: id}, nil
}

func (s *AuthService) LoginUser(ctx context.Context, user *pb.User) (*pb.AuthResponse, error) {
	email := user.GetEmail()
	pswd := user.GetPassword()
	id, err := s.storage.RegisterUser(ctx, email, pswd)
	if err != nil {
		s.logging.Errorf("Error getting chapters: %v", err)
		return nil, err
	}
	return &pb.AuthResponse{Id: id}, nil
}
