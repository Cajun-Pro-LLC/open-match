package main

import (
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
	"open-match.dev/open-match/pkg/matchfunction"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

const (
	matchName    = "match-function"
	maxTicketTTL = 600
)

var openMatchQueryService = fmt.Sprintf("%s:%s", os.Getenv("OM_QUERY_HOST"), os.Getenv("OM_QUERY_GRPC_PORT"))

type processor struct {
	client pb.QueryServiceClient
}

func (p *processor) Run(req *pb.RunRequest, stream pb.MatchFunction_RunServer) error {
	// Fetch tickets for the pools specified in the Match Profile.
	log.Printf("Generating proposals for function %v", req.GetProfile().GetName())

	poolTickets, err := matchfunction.QueryPools(stream.Context(), p.client, req.GetProfile().GetPools())
	if err != nil {
		log.Printf("Failed to query tickets for the given pools, got %s", err.Error())
		return err
	}
	var wg sync.WaitGroup
	expiredTickets := findExpiredTickets(poolTickets)
	if len(expiredTickets) > 0 {
		go func() {
			defer wg.Done()
			deleteTickets(expiredTickets)
		}()
	}

	// Generate proposals.
	proposals, err := findMatchProposals(req.GetProfile(), poolTickets)
	if err != nil {
		log.Printf("Failed to generate matches, got %s", err.Error())
		return err
	}

	log.Printf("Streaming %v proposals to Open Match", len(proposals))
	// Stream the generated proposals back to Open Match.
	for _, proposal := range proposals {
		if err := stream.Send(&pb.RunResponse{Proposal: proposal}); err != nil {
			log.Printf("Failed to stream proposals to Open Match, got %s", err.Error())
			return err
		}
	}
	wg.Wait()
	return nil
}

func createMatchProposal(poolTickets map[string][]*pb.Ticket, ticketsPerPoolPerMatch int) ([]*pb.Ticket, bool) {
	var matchTickets []*pb.Ticket
	insufficientTickets := false
	for pool, tickets := range poolTickets {
		if len(tickets) < ticketsPerPoolPerMatch {
			insufficientTickets = true
			break
		}
		// Remove the Tickets from this pool and add them to the match proposal.
		matchTickets = append(matchTickets, tickets[0:ticketsPerPoolPerMatch]...)
		poolTickets[pool] = tickets[ticketsPerPoolPerMatch:]
	}
	return matchTickets, insufficientTickets
}

func findMatchProposals(p *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error) {
	ticketsPerPoolPerMatch := 2
	var matches []*pb.Match
	count := 0
	for {
		matchTickets, insufficientTickets := createMatchProposal(poolTickets, ticketsPerPoolPerMatch)
		if insufficientTickets {
			break
		}
		matches = append(matches, &pb.Match{
			MatchId:       fmt.Sprintf("profile-%v-time-%v-%v", p.GetName(), time.Now().Format("2006-01-02T15:04:05.00"), count),
			MatchProfile:  p.GetName(),
			MatchFunction: matchName,
			Tickets:       matchTickets,
		})
		count++
	}
	return matches, nil
}

func findExpiredTickets(poolTickets map[string][]*pb.Ticket) []*pb.Ticket {
	var expiredTickets []*pb.Ticket
	for pool, tickets := range poolTickets {
		var validPoolTickets []*pb.Ticket
		for _, ticket := range tickets {
			if time.Now().After(ticket.GetCreateTime().AsTime().Add(time.Second * maxTicketTTL)) {
				expiredTickets = append(expiredTickets, ticket)
			} else {
				validPoolTickets = append(validPoolTickets, ticket)
			}
		}
		poolTickets[pool] = validPoolTickets
	}
	return expiredTickets
}

func deleteTickets(tickets []*pb.Ticket) {
	var wg sync.WaitGroup
	for _, ticket := range tickets {
		wg.Add(1)
		go func(ticket *pb.Ticket) {
			defer wg.Done()
			err := deleteTicket(ticket.GetId())
			if err != nil {
				fmt.Printf("Was not able to delete a ticket, err: %s\n", err.Error())
			}
		}(ticket)
	}

	wg.Wait()
}

func deleteTicket(ticketId string) error {
	r := regexp.MustCompile(`-(custom-frontend|mmf|director)-[a-z0-9]+-[a-z0-9]+$`)
	prefix := r.ReplaceAllString(os.Getenv("HOSTNAME"), "")
	underlined := strings.ReplaceAll(prefix, "-", "_")
	upper := strings.ToUpper(underlined)
	host := os.Getenv(fmt.Sprintf("%s_CUSTOM_FRONTEND_SVC_SERVICE_HOST", upper))
	url := fmt.Sprintf("http://%s:51504/v1/tickets/%s", host, ticketId)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if os.Getenv("SHOW_ENV") == "true" {
		fmt.Println("Environment Variables:")
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
	}

	fmt.Println("Starting MatchFunction Service...")

	s := grpc.NewServer()
	fmt.Printf("Creating grpc client for QueryService at %s\n", openMatchQueryService)
	conn, err := grpc.NewClient(openMatchQueryService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Could not dial Open Match Query Client service via gRPC, err: %v", err.Error()))
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
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("TCP net listener initialization failed for port %s, got %s", port, err.Error())
	}

	log.Printf("TCP net listener initialized for port %v", port)

	err = s.Serve(listener)
	if err != nil {
		log.Fatalf("gRPC serve failed, got %s", err.Error())
	}
}
