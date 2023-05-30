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

	client := pb.NewRankServiceClient(conn)
	req := pb.AddRankReq{UserId: 1, PhraseId: 2, Rank: 6, PaidRank: 7, Mp: "wb"}
	resp, err := client.AddRank(context.Background(), &req)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(resp)
}
