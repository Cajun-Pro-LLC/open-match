package main

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"open-match.dev/open-match/pkg/pb"
	"time"
)

func createMatchProposal(poolTickets map[string][]*pb.Ticket, ticketsPerPoolPerMatch int) ([]*pb.Ticket, bool) {
	l := log.With().Str("function", "createMatchProposal").Int("groupSize", ticketsPerPoolPerMatch).Logger()
	l.Trace().Msg("started")
	var matchTickets []*pb.Ticket
	for pool, tickets := range poolTickets {
		lp := l.With().Str("pool", pool).Int("tickets", len(tickets)).Logger()
		if len(tickets) < ticketsPerPoolPerMatch {
			lp.Trace().Msg("insufficient tickets for group size")
			continue
		}
		// Remove the Tickets from this pool and add them to the match proposal.
		matchTickets = append(matchTickets, tickets[0:ticketsPerPoolPerMatch]...)
		poolTickets[pool] = tickets[ticketsPerPoolPerMatch:]
		lp.Trace().Int("matchTickets", len(matchTickets)).Int("poolTickets", len(poolTickets[pool])).Msg("creating proposal")
		break
	}
	l.Trace().Int("matchTickets", len(matchTickets)).Bool("insufficient", len(matchTickets) == 0).Msg("finished")
	return matchTickets, len(matchTickets) == 0
}

func findMatchProposals(p *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error) {
	l := log.With().Str("function", "findMatchProposals").Str("profile", p.GetName()).Logger()
	ticketsPerPoolPerMatch := 2

	playerCountBytes, ok := p.GetExtensions()["playerCount"]
	if ok {
		var intValue wrapperspb.Int32Value
		err := proto.Unmarshal(playerCountBytes.GetValue(), &intValue)
		if err != nil {
			l.Err(err).Msg("failed to unmarshal playerCount")
		} else {
			ticketsPerPoolPerMatch = int(intValue.GetValue())
		}
	}
	l = l.With().Int("groupSize", ticketsPerPoolPerMatch).Logger()
	l.Trace().Msg("started")
	var matches []*pb.Match
	count := 0
	for {
		matchTickets, insufficientTickets := createMatchProposal(poolTickets, ticketsPerPoolPerMatch)
		l = l.With().Int("count", count).Int("matchTickets", len(matchTickets)).Bool("insufficient", insufficientTickets).Logger()
		if insufficientTickets {
			l.Trace().Msg("break")
			break
		}
		l.Trace().Msg("found match")
		matches = append(matches, &pb.Match{
			MatchId:       fmt.Sprintf("profile-%v-time-%v-%v", p.GetName(), time.Now().Format("2006-01-02T15:04:05.00"), count),
			MatchProfile:  p.GetName(),
			MatchFunction: matchName,
			Tickets:       matchTickets,
		})
		count++
	}
	l.Trace().Msg("finished")
	return matches, nil
}
