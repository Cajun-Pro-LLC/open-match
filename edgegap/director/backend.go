package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"open-match.dev/open-match/pkg/pb"
	"os"
	"strconv"
)

type Backend struct {
	client pb.BackendServiceClient
}

func NewBackend() (*Backend, error) {
	openMatchBackendService := fmt.Sprintf("%s:%s", os.Getenv("OM_BACKEND_HOST"), os.Getenv("OM_BACKEND_GRPC_PORT"))
	backend := &Backend{}
	conn, err := grpc.NewClient(openMatchBackendService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return backend, fmt.Errorf("error while communicating with Open Match Backend, err: %v", err.Error())
	}
	backend.client = pb.NewBackendServiceClient(conn)
	return backend, nil
}

// DeployGameserversForMatches deploys a game server with Edgegap API and assigns its IP to the match's tickets
func (b *Backend) DeployGameserversForMatches(matches []*pb.Match) []error {
	var errors []error
	for _, match := range matches {
		gs, err := deployMatchGameserver(match)
		if err != nil {
			errors = append(errors, fmt.Errorf("error while deploying gameserver for match %v, err: %v", match.GetMatchId(), err))
			// TODO::Tell the user they gotta requeue
			continue
		}
		if err := b.AssignMatch(gs); err != nil {
			errors = append(errors, fmt.Errorf("error while assigning match to gameserver for match %v, gameserver %v, err: %v", match.GetMatchId(), gs.Connection, err))
		}

		log.Printf("Assigned gameserver %v to match %v", gs.Connection, match.GetMatchId())
	}
	return nil
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
				TicketIds: gs.GetTicketIds(),
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
