package main

import (
	"github.com/cajun-pro-llc/open-match/utils"
	"open-match.dev/open-match/pkg/pb"
	"sync"
	"time"
)

func deployMatchGameserver(m *pb.Match) (*Gameserver, error) {
	// Creating API Client to communicate with arbitrium
	arbitrum := newArbitrum()
	gs := newGameServer(m)

	err := arbitrum.DeployGameserver(gs)
	if err != nil {
		return nil, err
	}

	return gs, nil
}

// deployGameserversForMatches deploys a game server with Edgegap API and assigns its IP to the match's tickets
func deployGameserversForMatches(matches []*pb.Match, backend *Backend) error {
	e := &NestedError{}
	for _, match := range matches {
		l := log.With().Str("matchId", match.GetMatchId()).Logger()
		gs, err := deployMatchGameserver(match)
		if err != nil {
			e.Add(err)
			l.Err(err).Msg("error while deploying gameserver for match")
			// TODO::Tell the user they gotta requeue
			continue
		}
		err = backend.AssignMatch(gs)
		if err != nil {
			e.Add(err)
			l.Err(err).Str("gameserver", gs.Connection).Msg("error while assigning match to gameserver")
		}

		l.Info().Str("gameserver", gs.Connection).Msg("assigned gameserver to match")
	}
	return e.Return()
}

// processMatches generates and assigns servers to matches for each profile
func processMatches(wg *sync.WaitGroup, profile *pb.MatchProfile, backend *Backend) {
	l := log.With().Str("profile", profile.GetName()).Logger()
	defer wg.Done()

	matches, err := backend.fetchMatchesForProfile(profile)
	if err != nil {
		l.Err(err).Msg("failed to fetch matches for profile")
		return
	}
	if len(matches) == 0 {
		l.Info().Msg("no matches generated")
	} else {
		l.Info().Int("count", len(matches)).Msg("generated matches")
	}

	err = deployGameserversForMatches(matches, backend)
	if err != nil {
		log.Err(err).Msg("errors occurred while deploying game servers")
	}
}

func main() {
	utils.LogEnv()
	log.Info().Msg("Starting service")
	arbitrum := newArbitrum()
	err := arbitrum.LoadConfiguration()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("Failed to load Arbitrum configuration")
	}
	matchmaker = arbitrum.matchmaker

	backend, err := NewBackend()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create backend")
	}
	for range time.Tick(time.Second * 5) {
		log.Debug().Msg("Creating matches...")
		var wg sync.WaitGroup
		for _, profile := range buildMatchmakerProfiles(matchmaker.Config.Profiles) {
			wg.Add(1)
			go processMatches(&wg, profile, backend)
		}
		wg.Wait()
	}
}
