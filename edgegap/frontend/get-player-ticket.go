package main

import (
	"context"
	"io"
	"open-match.dev/open-match/pkg/pb"
)

func getExistingTicket(playerId string) (*pb.Ticket, error) {
	service, conn := getQueryServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Err(closeErr).Msg("could not close query service client connection")
		}
	}()

	stream, err := service.QueryTickets(context.Background(), &pb.QueryTicketsRequest{Pool: &pb.Pool{}})
	if err != nil {
		return nil, err
	}

	var tickets []*pb.Ticket
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		tickets = append(tickets, resp.GetTickets()...)
	}
	for _, ticket := range tickets {
		if string(ticket.GetExtensions()["playerId"].GetValue()) == playerId {
			return ticket, nil
		}
	}
	return nil, nil
}
