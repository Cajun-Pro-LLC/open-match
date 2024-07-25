package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/anypb"
	"net/http"
	"open-match.dev/open-match/pkg/pb"
)

// Create a ticket by communicating with Open Match core Front End service
func createTicket(ctx echo.Context) error {
	c := ctx.(*Context)
	log.Info().Msg("Creating ticket")

	// Get The player IP. This will be used later to make a call at Arbitrium (Edgegap's solution)
	request := c.Request()
	playerIP := c.Echo().IPExtractor(request)

	userTicketRequest := CreateTicketRequest{}

	// Bind the request JSON body to our model
	err := c.Bind(&userTicketRequest)
	if err != nil {
		log.Err(err).Msg("request did not match CreateTicketRequest attributes")
		return c.RespondError(http.StatusBadRequest)
	}
	log.Debug().Object("request", userTicketRequest).Msg("payload")

	service, conn := getFrontendServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Err(err).Msg("could not close frontend connection")
		}
	}()

	searchFields := &pb.SearchFields{
		StringArgs: userTicketRequest.Matchmaking.Selectors,
		DoubleArgs: userTicketRequest.Matchmaking.Filters,
		Tags:       []string{userTicketRequest.ProfileId},
	}
	// searchFields.StringArgs["playerId"] = userTicketRequest.PlayerId

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
	var ticket *pb.Ticket
	ticket, err = getExistingTicket(userTicketRequest.PlayerId)
	if err != nil {
		log.Err(err).Msg("could not search for existing ticket")
	}
	if ticket != nil {
		log.Debug().Str("ticketId", ticket.GetId()).Msg("ticket already exists, reusing existing ticket")
		return c.Respond(ticket)
	}

	ticket, err = service.CreateTicket(context.Background(), req)
	if err != nil {
		log.Err(err).Msg("could not create a ticket")
		return c.RespondErrorCustom(http.StatusInternalServerError, err.Error())
	}

	return c.Respond(ticket)
}
