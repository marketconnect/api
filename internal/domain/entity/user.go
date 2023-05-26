package entity

import pb "github.com/marketconnect/contracts/pb/auth/v1"

type User struct {
	Username string `json:"username" validate:"required,min=5,max=30"`
	Password string `json:"password" validate:"required,min=8,max=50"`
	Email    string `json:"email" validate:"required,email"`
}

func (user *User) ToGrpcUser() *pb.User {
	return &pb.User{
		Email:    user.Email,
		Password: user.Password,
	}
}
