package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"open-match.dev/open-match/pkg/pb"
	"os"
	"strconv"
)

type Backend struct {
	client pb.BackendServiceClient
}

func NewBackend() (*Backend, error) {
	omBackendUri := os.Getenv("OM_BACKEND_HOST") + ":" + os.Getenv("OM_BACKEND_GRPC_PORT")
	l := log.With().Str("uri", omBackendUri).Logger()
	l.Debug().Msg("creating grpc client for backend service")
	conn, err := grpc.NewClient(omBackendUri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Err(err).Msg("could not create grpc client for backend service")
		return nil, err
	}
	backend := &Backend{
		client: pb.NewBackendServiceClient(conn),
	}
	return backend, nil
}

// Fetch profile's matches
func (b *Backend) fetchMatchesForProfile(p *pb.MatchProfile) ([]*pb.Match, error) {
	// Making request object
	port, _ := strconv.Atoi(os.Getenv("OM_MMF_GRPC_PORT"))
	req := &pb.FetchMatchesRequest{
		Config: &pb.FunctionConfig{
			Host: os.Getenv("OM_MMF_HOST"),
			Port: int32(port),
			Type: pb.FunctionConfig_GRPC,
		},
		Profile: p,
	}

	// Getting match proposals
	stream, err := b.client.FetchMatches(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var result []*pb.Match
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		result = append(result, resp.GetMatch())
	}

	return result, nil
}

func (b *Backend) AssignMatch(gs *Gameserver) error {
	req := &pb.AssignTicketsRequest{
		Assignments: []*pb.AssignmentGroup{
			{
				TicketIds: gs.Players().Tickets(),
				Assignment: &pb.Assignment{
					Connection: gs.Connection,
				},
			},
		},
	}
	_, err := b.client.AssignTickets(context.Background(), req)
	if err != nil {
		return err
	}

	return nil
}
