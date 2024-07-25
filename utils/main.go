package utils

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCClient(uri string) (*grpc.ClientConn, error) {
	return grpc.NewClient(uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
