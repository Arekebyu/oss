package main

import (
	"context"
	"fmt"
	"log"
	pb "oss/pb"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect %v", err)
	}
	defer conn.Close()
	client := pb.NewMLServiceClient(conn)

	req := &pb.RankRequest{
		Query: "Reshape a tensor",
		Candidates: []*pb.Document{
			{
				Id:             "doc1",
				Title:          "Pytorch View Documentation",
				ContentSnippet: "The view method returns a new tensor with the same data as the self tensor but of a different shape.",
			},
			{
				Id:             "doc2",
				Title:          "React Router Guide",
				ContentSnippet: "React Router enables client side routing. It is not related to tensors.",
			},
			{
				Id:             "doc3",
				Title:          "Numpy Reshape",
				ContentSnippet: "Gives a new shape to an array without changing its data.",
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	fmt.Println("requesting ml service")

	r, err := client.ReRank(ctx, req)
	if err != nil {
		log.Fatalf("ranking failed, %v", err)
	}

	fmt.Printf("Page rank success, %d \n", len(r.GetResults()))
	for _, res := range r.GetResults() {
		fmt.Printf("Doc ID %s | Score %.4f\n", res.GetId(), res.GetScore())
	}
}
