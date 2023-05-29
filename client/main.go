package main

import (
	"context"
	"fmt"
	"log"

	pb "mc_api/pkg/api"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:1953", grpc.WithInsecure())
	if err != nil {
		log.Println(err)
	}

	client := pb.NewAuthServiceClient(conn)
	resp, err := client.LoginUser(context.Background(), &pb.User{Email: "aassfd@mail.ru", Password: "111222333444"})
	if err != nil {
		log.Println(err)
	}

	fmt.Println(resp)
}
