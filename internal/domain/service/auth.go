package service

import (
	"context"
	"fmt"

	"api/internal/domain/entity"

	pb "github.com/marketconnect/contracts/pb/auth/v1"
)

type authService struct {
	client pb.AuthServiceClient
}

func NewAuthUsecase(client pb.AuthServiceClient) *authService {
	return &authService{client: client}
}

func (u authService) Login(ctx context.Context, user entity.User) (uint64, error) {
	grpcUser := user.ToGrpcUser()
	resp, err := u.client.LoginUser(ctx, grpcUser)

	if err != nil {
		return 0, fmt.Errorf("WRONG WRONG WRONG")
	}
	fmt.Printf("Service Login %s ---> %s ---> %s", user.Password, user.Username, user.Email)
	return resp.Id, nil

}

func (u authService) Register(ctx context.Context, user entity.User) (uint64, error) {
	grpcUser := user.ToGrpcUser()
	resp, err := u.client.RegisterUser(ctx, grpcUser)

	if err != nil {
		return 0, fmt.Errorf("WRONG WRONG WRONG")
	}
	fmt.Printf("Service Register %s ---> %s ---> %s", user.Password, user.Username, user.Email)
	return resp.Id, nil
}
