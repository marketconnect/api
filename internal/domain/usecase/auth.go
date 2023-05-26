package usecase

import (
	"context"
	"fmt"

	"api/internal/domain/entity"
)

type authUsecase struct {
}

func NewAuthUsecase() *authUsecase {
	return &authUsecase{}
}

func (u authUsecase) Login(ctx context.Context, user entity.User) bool {
	if (user.Email != "ivan@gmail.com") || (user.Password != "111222333") || (user.Username != "IvanB") {
		return false
	}
	fmt.Printf("Usecase %s ---> %s", user.Password, user.Username)
	return true
}

func (u authUsecase) Register(ctx context.Context, user entity.User) error {
	if false {
		return fmt.Errorf("")
	}
	fmt.Printf("Usecase Register %s ---> %s ---> %s", user.Password, user.Username, user.Email)
	return nil
}
