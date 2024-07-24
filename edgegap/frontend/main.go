package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"log"
	"net/http"
	"open-match.dev/open-match/pkg/pb"
	"os"
)

// TicketRequestModel represent the model that we should receive for our create ticket endpoint
type TicketRequestModel struct {
	ProfileId   string `json:"edgegap_profile_id"`
	PlayerId    string `json:"player_id"`
	Matchmaking struct {
		Selectors map[string]string  `json:"selector_data"`
		Filters   map[string]float64 `json:"filter_data"`
	} `json:"matchmaking_data"`
}

// Create a ticket by communicating with Open Match core Front End service
func createTicket(ctx echo.Context) error {
	c := ctx.(*Context)
	log.Println("Creating ticket...")

	// Get The player IP. This will be used later to make a call at Arbitrium (Edgegap's solution)
	request := c.Request()
	playerIP := c.Echo().IPExtractor(request)

	userTicketRequest := TicketRequestModel{}

	// Bind the request JSON body to our model
	err := c.Bind(&userTicketRequest)
	if err != nil {
		log.Println("Request Payload didn't match TicketRequestModel attributes")
		return c.RespondError(http.StatusBadRequest)
	}
	tReq, _ := json.Marshal(userTicketRequest)
	log.Printf("Request Payload: %s", string(tReq))

	service, conn := getFrontendServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing frontend connection: %v\n", closeErr.Error())
		}
	}()

	searchFields := &pb.SearchFields{
		StringArgs: userTicketRequest.Matchmaking.Selectors,
		DoubleArgs: userTicketRequest.Matchmaking.Filters,
		Tags:       []string{userTicketRequest.ProfileId},
	}
	searchFields.StringArgs["playerId"] = userTicketRequest.PlayerId

	sf, _ := json.Marshal(searchFields)
	log.Printf("Search Fields: %s", string(sf))

	req := &pb.CreateTicketRequest{
		Ticket: &pb.Ticket{
			SearchFields: searchFields,
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
		return c.Respond(existingTicket)
	}

	ticket, err := service.CreateTicket(context.Background(), req)
	if err != nil {
		log.Printf("Was not able to create a ticket, err: %s\n", err.Error())
		return c.RespondErrorCustom(http.StatusInternalServerError, err.Error())
	}

	return c.Respond(ticket)
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

func getTicket(ctx echo.Context) error {
	c := ctx.(*Context)
	ticketID := c.Param("ticketId")

	service, conn := getFrontendServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing frontend connection: %v\n", closeErr.Error())
		}
	}()

	ticket, err := service.GetTicket(context.Background(), &pb.GetTicketRequest{TicketId: ticketID})
	if err != nil {
		log.Printf("Was not able to get a ticket, err: %s\n", err.Error())
		return c.RespondErrorCustom(http.StatusNotFound, "Ticket not found")
	}

	return c.Respond(ticket)
}

func deleteTicket(ctx echo.Context) error {
	c := ctx.(*Context)
	ticketID := c.Param("ticketId")

	service, conn := getFrontendServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing frontend connection: %v\n", closeErr.Error())
		}
	}()

	_, err := service.DeleteTicket(context.Background(), &pb.DeleteTicketRequest{TicketId: ticketID})
	if err != nil {
		fmt.Printf("Was not able to delete a ticket, err: %s\n", err.Error())
		return c.RespondErrorCustom(http.StatusNotFound, "Ticket not found")
	}

	return c.Respond(pb.Ticket{Id: ticketID})
}
func main() {
	if os.Getenv("SHOW_ENV") == "true" {
		fmt.Println("Environment Variables:")
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
	}
	fmt.Println("Starting Frontend Service...")

	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(&Context{c})
		}
	})
	e.Use(middleware.RequestID())
	// How to extract IP
	e.IPExtractor = echo.ExtractIPDirect()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Open Match Frontend")
	})

	v1 := e.Group("/v1")

	tickets := v1.Group("/tickets")
	// Create a ticket
	tickets.POST("", createTicket)
	// Get a ticket
	tickets.GET("/:ticketId", getTicket)
	// Delete a ticket
	tickets.DELETE("/:ticketId", deleteTicket)

	// Serve on the edgegap environment variable defined port
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("HTTP_SERVE_PORT"))))
}
