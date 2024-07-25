package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"open-match.dev/open-match/pkg/pb"
)

func deleteTicket(ctx echo.Context) error {
	c := ctx.(*Context)
	ticketId := c.Param("ticketId")

	service, conn := getFrontendServiceClient()
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Err(closeErr).Msg("could not close frontend service client connection")
		}
	}()

	_, err := service.DeleteTicket(context.Background(), &pb.DeleteTicketRequest{TicketId: ticketId})
	if err != nil {
		log.Err(err).Str("ticketId", ticketId).Msg("could not delete ticket")
		return c.RespondErrorCustom(http.StatusNotFound, "Ticket not found")
	}

	return c.Respond(pb.Ticket{Id: ticketId})
}
