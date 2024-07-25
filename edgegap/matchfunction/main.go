package main

import (
	"github.com/cajun-pro-llc/open-match/utils"
	"google.golang.org/grpc"
	"net"
	"open-match.dev/open-match/pkg/pb"
	"os"
)

func main() {
	utils.LogEnv()
	log.Info().Msg("Starting service")

	s := grpc.NewServer()
	openMatchQueryService := os.Getenv("OM_QUERY_HOST") + ":" + os.Getenv("OM_QUERY_GRPC_PORT")
	log.Info().Str("uri", openMatchQueryService).Msg("creating grpc client for query service")
	conn, err := utils.NewGRPCClient(openMatchQueryService)
	if err != nil {
		log.Panic().Err(err).Msg("could not dial open-match query client service via grpc")
	}

	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing queryClient connection: %v\n", closeErr.Error())
		}
	}()

	// Register Query Service Server & Match Function Server
	client := pb.NewQueryServiceClient(conn)
	pb.RegisterQueryServiceServer(s, &server{
		client: client,
	})
	pb.RegisterMatchFunctionServer(s, &processor{
		client: client,
	})

	port := os.Getenv("GRPC_SERVE_PORT") // defaults to 50502
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal().Err(err).Str("port", port).Msg("tcp net listener initialization failed")
	}

	log.Info().Str("port", port).Msg("TCP net listener initialized for port")

	err = s.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("grpc serve failed")
	}
}
