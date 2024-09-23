package main

import (
	"fmt"
	"sync"

	"open-match.dev/open-match/pkg/matchfunction"
	"open-match.dev/open-match/pkg/pb"
	"runtime/debug"
)

type processor struct {
	client pb.QueryServiceClient
}

func (p *processor) Run(req *pb.RunRequest, stream pb.MatchFunction_RunServer) error {
	// Fetch tickets for the pools specified in the Match Profile.
	l := log.With().Str("profile", req.GetProfile().GetName()).Logger()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic: \n", string(debug.Stack()))
		}
	}()

	l.Debug().Msg("processing proposals")
	poolTickets, err := matchfunction.QueryPools(stream.Context(), p.client, req.GetProfile().GetPools())
	if err != nil {
		l.Printf("Failed to query tickets for the given pools, got %s\n", err.Error())
		return err
	}
	for pool, poolTicket := range poolTickets {
		l.Debug().Str("pool", pool).Int("count", len(poolTicket)).Msg("processing tickets")
	}
	var wg sync.WaitGroup
	expiredTickets := findExpiredTickets(poolTickets)
	if len(expiredTickets) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deleteTickets(expiredTickets)
		}()
	}

	wg.Wait()

	// Generate proposals.
	proposals, err := findMatchProposals(req.GetProfile(), poolTickets)
	if err != nil {
		log.Err(err).Msg("could not generate matches")
		return err
	}

	if len(proposals) > 0 {
		log.Info().Int("count", len(proposals)).Msg("proposals generated")
	} else {
		log.Info().Msg("no proposals generated")
	}
	// Stream the generated proposals back to Open Match.
	for _, proposal := range proposals {
		if err := stream.Send(&pb.RunResponse{Proposal: proposal}); err != nil {
			log.Err(err).Msg("could not stream proposals to open-match")
			return err
		}
	}
	return nil
}
