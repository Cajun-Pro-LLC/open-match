package main

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"open-match.dev/open-match/pkg/pb"
	"os"
	"regexp"
	"strings"
)

var openMatchFrontendService = fmt.Sprintf("%s:%s", os.Getenv("OM_FRONTEND_HOST"), os.Getenv("OM_FRONTEND_GRPC_PORT"))

// Get an object that can communicate with OpenMatch Frontend service.
func getFrontendServiceClient() (pb.FrontendServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(openMatchFrontendService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil
	}
	return pb.NewFrontendServiceClient(conn), conn
}

// Get an object that can communicate with OpenMatch Query service.
func getQueryServiceClient() (pb.QueryServiceClient, *grpc.ClientConn) {
	r := regexp.MustCompile(`-(custom-frontend|mmf|director)-[a-z0-9]+-[a-z0-9]+$`)
	underlined := strings.ReplaceAll(os.Getenv("HOSTNAME"), "-", "_")
	prefix := r.ReplaceAllString(underlined, "")
	upper := strings.ToUpper(prefix)
	httpPort := os.Getenv(fmt.Sprintf("%s_MMF_SVC_SERVICE_PORT_HTTP", upper))
	host := os.Getenv(fmt.Sprintf("%s_MMF_SVC_PORT_%s_TCP_ADDR", upper, httpPort))
	port := os.Getenv(fmt.Sprintf("%s_MMF_SVC_SERVICE_PORT_GRPC", upper))
	openMatchQueryService := fmt.Sprintf("%s:%s", host, port)
	conn, err := grpc.NewClient(openMatchQueryService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Could not dial Open Match Query Client service via gRPC, err: %v", err.Error()))
	}
	return pb.NewQueryServiceClient(conn), conn
}
