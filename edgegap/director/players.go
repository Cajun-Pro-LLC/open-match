package main

import "open-match.dev/open-match/pkg/pb"

type playerDetails struct {
	PlayerId string
	IP       string
	TicketId string
}

func newPlayerDetails(t *pb.Ticket) *playerDetails {
	return &playerDetails{
		PlayerId: string(t.Extensions["playerId"].Value),
		IP:       string(t.Extensions["playerIp"].Value),
		TicketId: t.Id,
	}
}
