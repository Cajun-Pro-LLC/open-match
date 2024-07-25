package main

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"open-match.dev/open-match/pkg/pb"
	"time"
)

func createMatchProposal(poolTickets map[string][]*pb.Ticket, ticketsPerPoolPerMatch int) ([]*pb.Ticket, bool) {
	var matchTickets []*pb.Ticket
	insufficientTickets := false
	for pool, tickets := range poolTickets {
		log.Debug().Str("function", "createMatchProposal").Str("pool", pool).Int("tickets", len(tickets)).Int("groupSize", ticketsPerPoolPerMatch).Msg("creating match proposals")
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

	playerCountBytes, ok := p.GetExtensions()["playerCount"]
	if ok {
		var intValue wrapperspb.Int32Value
		err := proto.Unmarshal(playerCountBytes.GetValue(), &intValue)
		if err != nil {
			log.Err(err).Msg("failed to unmarshal playerCount")
		} else {
			ticketsPerPoolPerMatch = int(intValue.GetValue())
		}
	}

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
