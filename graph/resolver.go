package graph

import (
	"fmt"
	"os"

	"github.com/himanshu-holmes/social-feed-system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

const defaultPort = "6001"
type Resolver struct{
	PostClient proto.TimelineServiceClient
}
func  NewResolver()*Resolver{
	port := os.Getenv("GRPC_PORT")
		if port == "" {
			port = defaultPort
		}
	postConn,err := grpc.NewClient(fmt.Sprintf(":%s",port),grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Error while connecting to post service",err)
		panic(err)
	}
	postClient := proto.NewTimelineServiceClient(postConn)
	return &Resolver{
		PostClient: postClient,
	}
}
