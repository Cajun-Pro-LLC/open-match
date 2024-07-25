package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"open-match.dev/open-match/pkg/pb"
)

func getTicket(ctx echo.Context) error {
	c := ctx.(*Context)
	ticketId := c.Param("ticketId")

	service, conn := getFrontendServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Err(closeErr).Msg("could not close frontend service client connection")
		}
	}()

	ticket, err := service.GetTicket(context.Background(), &pb.GetTicketRequest{TicketId: ticketId})
	if err != nil {
		log.Err(err).Str("ticketId", ticketId).Msg("could not find matching ticket")
		return c.RespondErrorCustom(http.StatusNotFound, "Ticket not found")
	}

	return c.Respond(ticket)
}
