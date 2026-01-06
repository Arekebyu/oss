package main

import (
	"log"
	"oss/internal/api"
	"oss/internal/config"
	"oss/internal/search"
	pb "oss/pb"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.LoadConfig()
	// connect to python
	var pyML *grpc.ClientConn
	var err error
	for i := 0; i < 10; i++ {
		pyML, err = grpc.NewClient(cfg.MLServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		}
		log.Printf("Waiting for DB, attempt %d", (i+1))
		time.Sleep(5 * time.Second)
	}
	MLClient := pb.NewMLServiceClient(pyML)
	defer pyML.Close()

	// connect to elasticsearch
	es, err := search.NewClient(cfg.ElasticsearchURL)
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

	log.Printf("server running on port %s\n", cfg.Port)
	err = r.Run(":" + cfg.Port)
	if err != nil {
		log.Fatal(err)
	}
}
