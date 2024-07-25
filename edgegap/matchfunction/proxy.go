package main

import (
	"context"
	"io"
	"open-match.dev/open-match/pkg/pb"
)

// gRPC Query Service Server acting as a Proxy
type server struct {
	pb.UnimplementedQueryServiceServer
	client pb.QueryServiceClient
}

// QueryTickets updates the function of the same name to the backend.
func (s *server) QueryTickets(req *pb.QueryTicketsRequest, srv pb.QueryService_QueryTicketsServer) error {
	// Query the actual service with the incoming request
	stream, err := s.client.QueryTickets(context.Background(), req)
	if err != nil {
		log.Printf("Error querying tickets: %v\n", err)
		// You can return the error back to the original caller
		return err
	}

	// Stream the response back to the caller
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Streaming error: %v\n", err)
			return err
		}
		if err := srv.Send(resp); err != nil {
			log.Printf("Error sending to client: %v\n", err)
			return err
		}
	}

	return nil
}

// QueryTicketIds gets the list of TicketIDs that meet all the filtering criteria requested by the pool.
func (s *server) QueryTicketIds(req *pb.QueryTicketIdsRequest, srv pb.QueryService_QueryTicketIdsServer) error {
	// Calling the actual Query Service with same request
	stream, err := s.client.QueryTicketIds(context.Background(), req)
	if err != nil {
		log.Printf("Error querying ticketIds: %v\n", err)
		// you can return the error back to the original caller
		return err
	}

	// Stream the response back to the caller
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Streaming error: %v\n", err)
			return err
		}
		if err := srv.Send(resp); err != nil {
			log.Printf("Error sending to client: %v\n", err)
			return err
		}
	}

	return nil
}

// QueryBackfills gets a list of Backfills.
func (s *server) QueryBackfills(req *pb.QueryBackfillsRequest, srv pb.QueryService_QueryBackfillsServer) error {
	// Calling the actual Query Service with same request
	stream, err := s.client.QueryBackfills(context.Background(), req)
	if err != nil {
		log.Printf("Error querying backfills: %v\n", err)
		// you can return the error back to the original caller
		return err
	}

	// Stream the response back to the caller
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Streaming error: %v\n", err)
			return err
		}
		if err := srv.Send(resp); err != nil {
			log.Printf("Error sending to client: %v\n", err)
			return err
		}
	}

	return nil
}
