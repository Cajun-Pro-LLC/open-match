package main

import (
	"fmt"
	"github.com/cajun-pro-llc/open-match/utils"
	"google.golang.org/protobuf/types/known/anypb"
	"open-match.dev/open-match/pkg/pb"
	"strings"
	"time"
)

func createStaticMatchProposal(poolTickets map[string][]*pb.Ticket, playerCount int) ([]*pb.Ticket, bool) {
	l := log.With().Str("function", "createStaticMatchProposal").Int("playerCount", playerCount).Logger()
	l.Trace().Msg("started")
	var matchTickets []*pb.Ticket
	for pool, tickets := range poolTickets {
		lp := l.With().Str("pool", pool).Int("tickets", len(tickets)).Logger()
		if len(tickets) < playerCount {
			lp.Trace().Msg("insufficient tickets for group size")
			continue
		}
		// Remove the Tickets from this pool and add them to the match proposal.
		matchTickets = append(matchTickets, tickets[0:playerCount]...)
		poolTickets[pool] = tickets[playerCount:]
		lp.Trace().Int("matchTickets", len(matchTickets)).Int("poolTickets", len(poolTickets[pool])).Msg("creating proposal")
		break
	}
	l.Trace().Int("matchTickets", len(matchTickets)).Bool("insufficient", len(matchTickets) == 0).Msg("finished")
	return matchTickets, len(matchTickets) == 0
}

func createSteppedMatchProposal(poolTickets map[string][]*pb.Ticket, playerMin int, playerMax int, playerStep int) ([]*pb.Ticket, bool) {
	l := log.With().Str("function", "createSteppedMatchProposal").Int("playerMin", playerMin).Int("playerMax", playerMax).Int("playerStep", playerStep).Logger()
	l.Trace().Msg("started")
	var matchTickets []*pb.Ticket
	insufficient := true

	stepDuration := time.Duration(playerStep) * time.Second // convert step in seconds into time.Duration

	// Current time when function is called
	currentTime := time.Now()

	for playerCount := playerMax; playerCount >= playerMin; playerCount-- {
		lp := l.With().Int("playerCount", playerCount).Logger()
		lp.Trace().Msg("trying to find match with current player count")

		// Create a new map for recently created tickets
		recentPoolTickets := make(map[string][]*pb.Ticket)

		// Filter tickets: ticket is considered eligible if it is created earlier, or within the last stepDuration
		for pool, tickets := range poolTickets {
			for _, ticket := range tickets {
				if currentTime.Sub(ticket.CreateTime.AsTime()) >= stepDuration {
					recentPoolTickets[pool] = append(recentPoolTickets[pool], ticket)
				}
			}
		}

		matchTickets, insufficient = createStaticMatchProposal(recentPoolTickets, playerCount)
		if !insufficient {
			break
		}

		lp.Trace().Msg("insufficient tickets for current player count")
		// subtract step duration for the next iteration
		currentTime = currentTime.Add(-stepDuration)
	}
	l.Trace().Int("matchTickets", len(matchTickets)).Bool("insufficient", insufficient).Msg("finished")
	return matchTickets, insufficient
}

func findMatchProposals(p *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error) {
	l := log.With().Str("function", "findMatchProposals").Str("profile", p.GetName()).Logger()
	playerCount := int(utils.ProtoToInt32(p.GetExtensions()["playerCount"], 2))
	playerMin := int(utils.ProtoToInt32(p.GetExtensions()["playerMin"], 2))
	playerMax := int(utils.ProtoToInt32(p.GetExtensions()["playerMax"], 5))
	playerStep := int(utils.ProtoToInt32(p.GetExtensions()["playerStep"], -1))
	isStepped := strings.Contains(strings.ToLower(p.GetName()), "ffa") && playerStep > -1
	l = l.With().Int("groupSize", playerCount).Logger()
	l.Trace().Msg("started")
	var matches []*pb.Match
	count := 0
	for {
		var matchTickets []*pb.Ticket
		var insufficientTickets bool
		if isStepped {
			matchTickets, insufficientTickets = createSteppedMatchProposal(poolTickets, playerMin, playerMax, playerStep)
		} else {
			matchTickets, insufficientTickets = createStaticMatchProposal(poolTickets, playerCount)
		}
		l = l.With().Int("count", count).Int("matchTickets", len(matchTickets)).Bool("insufficient", insufficientTickets).Logger()
		if insufficientTickets {
			l.Trace().Msg("break")
			break
		}
		l.Trace().Msg("found match")
		extensions := map[string]*anypb.Any{
			"isStepped": {
				TypeUrl: "type.googleapis.com/google.protobuf.BoolValue",
				Value:   utils.BoolToProto(isStepped),
			},
		}
		for key, value := range matchTickets[0].SearchFields.StringArgs {
			extensions[key] = &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
				Value:   []byte(value),
			}
		}
		if isStepped {
			extensions["playerCount"] = &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
				Value:   utils.Int32ToProto(int32(len(matchTickets))),
			}
		}
		matches = append(matches, &pb.Match{
			MatchId:       fmt.Sprintf("profile-%v-time-%v-%v", p.GetName(), time.Now().Format(time.RFC3339), count),
			MatchProfile:  p.GetName(),
			MatchFunction: matchName,
			Tickets:       matchTickets,
			Extensions:    extensions,
		})
		count++
	}
	l.Trace().Msg("finished")
	return matches, nil
}
