package main

import (
	"log"
	"oss/internal/api"
	"oss/internal/search"
	pb "oss/pb"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// connect to python
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("python service did not connect %v", err)
	}
	defer conn.Close()
	MLClient := pb.NewMLServiceClient(conn)

	// connect to elasticsearch
	es, err := search.NewClient("http://localhost:9200")
	if err != nil {
		log.Fatalf("elasticsearch could not connect %v", err)
	}

	svc := &api.SearchService{
		ESClient: es,
		MLClient: MLClient,
	}
	handler := &api.Handler{Service: svc}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	r.GET("search", handler.HandleSearch)

	log.Println("server running on https://localhost:8080")
	err = r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
