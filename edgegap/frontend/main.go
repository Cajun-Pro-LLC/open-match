package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"log"
	"net/http"
	"open-match.dev/open-match/pkg/pb"
)

const (
	openMatchFrontendService = "open-match-frontend:51504"
	openMatchQueryService    = "open-match-query-client:12345"
)

// TicketRequestModel represent the model that we should receive for our create ticket endpoint
type TicketRequestModel struct {
	Category string
	Mode     string
	PlayerId string
}

// Create a ticket by communicating with Open Match core Front End service
func createTicket(echoContext echo.Context) error {
	log.Println("Creating ticket...")

	// Get The player IP. This will be used later to make a call at Arbitrium (Edgegap's solution)
	echoServer := echoContext.Echo()
	request := echoContext.Request()

	playerIP := echoServer.IPExtractor(request)

	userTicketRequest := TicketRequestModel{}

	// Bind the request JSON body to our model
	err := echoContext.Bind(&userTicketRequest)

	if err != nil {
		panic("Request Payload didn't match TicketRequestModel attributes")
	}

	service, conn := getFrontendServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing frontend connection: %v\n", closeErr.Error())
		}
	}()

	req := &pb.CreateTicketRequest{
		Ticket: &pb.Ticket{
			SearchFields: &pb.SearchFields{
				// Tags can support multiple values but for simplicity, the demo function
				// assumes only single mode selection per Ticket.
				Tags: []string{
					userTicketRequest.Category,
					userTicketRequest.Mode,
				},
				StringArgs: map[string]string{"playerId": userTicketRequest.PlayerId},
			},
			Extensions: map[string]*anypb.Any{
				// Adding player IP to create the game server later using Arbitrium (Edgegap's solution)
				// You can add other values in extensions. Those values will be ignored by Open Match. They are meant tu use by the developer.
				// Find all valid type here: https://developers.google.com/protocol-buffers/docs/reference/google.protobuf
				"playerIp": {
					TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
					Value:   []byte(playerIP),
				},
				"playerId": {
					TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
					Value:   []byte(userTicketRequest.PlayerId),
				},
			},
		},
	}

	existingTicket, err := getExistingTicket(userTicketRequest.PlayerId)
	if err != nil {
		log.Printf("Error checking for existing ticket: %v", err.Error())
	}
	if existingTicket != nil {
		return echoContext.JSON(http.StatusOK, existingTicket)
	}

	ticket, err := service.CreateTicket(context.Background(), req)
	if err != nil {
		log.Printf("Was not able to create a ticket, err: %s\n", err.Error())
	}

	return echoContext.JSON(http.StatusOK, ticket)
}

func getExistingTicket(playerId string) (*pb.Ticket, error) {
	service, conn := getQueryServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing query service client connection: %v\n", closeErr.Error())
		}
	}()
	req := &pb.QueryTicketsRequest{
		Pool: &pb.Pool{
			StringEqualsFilters: []*pb.StringEqualsFilter{
				{StringArg: "playerId", Value: playerId},
			},
		},
	}
	stream, err := service.QueryTickets(context.Background(), req)
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
	if len(tickets) > 0 {
		if len(tickets) > 1 {
			log.Printf("expected 0-1 tickets for %v, found %d.", playerId, len(tickets))
		}
		for _, ticket := range tickets {
			return ticket, nil
		}
	}
	return nil, nil
}

// Get an object that can communicate with Open Match Front End service.
func getFrontendServiceClient() (pb.FrontendServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(openMatchFrontendService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Could not dial Open Match Frontend service via gRPC, err: %v", err.Error()))
	}

	return pb.NewFrontendServiceClient(conn), conn
}

// Get an object that can communicate with Open Match Front End service.
func getQueryServiceClient() (pb.QueryServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(openMatchQueryService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Could not dial Open Match Query Client service via gRPC, err: %v", err.Error()))
	}
	return pb.NewQueryServiceClient(conn), conn
}

func getTicket(echoContext echo.Context) error {
	ticketID := echoContext.Param("ticketId")

	service, conn := getFrontendServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing frontend connection: %v\n", closeErr.Error())
		}
	}()

	req := &pb.GetTicketRequest{
		TicketId: ticketID,
	}

	ticket, err := service.GetTicket(context.Background(), req)
	if err != nil {
		log.Printf("Was not able to get a ticket, err: %s\n", err.Error())
		return echo.NewHTTPError(http.StatusNotFound, "Resource not found")
	}

	return echoContext.JSON(http.StatusOK, ticket)
}

func deleteTicket(echoContext echo.Context) error {
	ticketID := echoContext.Param("ticketId")

	service, conn := getFrontendServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing frontend connection: %v\n", closeErr.Error())
		}
	}()

	req := &pb.DeleteTicketRequest{
		TicketId: ticketID,
	}

	_, err := service.DeleteTicket(context.Background(), req)

	if err != nil {
		fmt.Printf("Was not able to delete a ticket, err: %s\n", err.Error())
		return echo.NewHTTPError(http.StatusNotFound, "Resource not found")
	}

	return echoContext.JSON(http.StatusOK, pb.Ticket{Id: ticketID})
}

func main() {
	fmt.Println("Starting Frontend Service...")

	e := echo.New()

	// How to extract IP
	e.IPExtractor = echo.ExtractIPDirect()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Open Match Frontend")
	})

	v1 := e.Group("/v1")

	tickets := v1.Group("/tickets")
	// Create a ticket
	tickets.POST("/", createTicket)
	// Get a ticket
	tickets.GET("/:ticketId", getTicket)
	// Delete a ticket
	tickets.DELETE("/:ticketId", deleteTicket)

	e.Logger.Fatal(e.Start(":51504"))
}
