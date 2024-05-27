package main

import (
	"sync"

	"fmt"
	"log"
	"open-match.dev/open-match/pkg/pb"
	"time"
)

const (
	openMatchMatchFunctionHost = "match-function"
	openMatchMatchFunctionPort = 50502
	openMatchBackendService    = "open-match-backend:50505"
	// Game server data
	gameServerPort = "gameport"
	appName        = "gameserver"
	appVersion     = "v1.0.0"
	arbitriumAPI   = "https://staging-api.edgegap.com/"
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
func deployGameserversForMatches(matches []*pb.Match, backend *Backend) []error {
	var errors []error
	for _, match := range matches {
		gs, err := deployMatchGameserver(match)
		if err != nil {
			errors = append(errors, fmt.Errorf("error while deploying gameserver for match %v, err: %v", match.GetMatchId(), err))
			// TODO::Tell the user they gotta requeue
			continue
		}
		if err := backend.AssignMatch(gs); err != nil {
			errors = append(errors, fmt.Errorf("error while assigning match to gameserver for match %v, gameserver %v, err: %v", match.GetMatchId(), gs.Connection, err))
		}

		log.Printf("Assigned gameserver %v to match %v", gs.Connection, match.GetMatchId())
	}
	return nil
}

// processMatches generates and assigns servers to matches for each profile
func processMatches(wg *sync.WaitGroup, profile *pb.MatchProfile, backend *Backend) {
	defer wg.Done()
	matches, err := backend.fetchMatchesForProfile(profile)
	if err != nil {
		log.Printf("Failed to fetch matches for profile %v, got %s", profile.GetName(), err.Error())
		return
	}
	log.Printf("Generated %v matches for profile %v", len(matches), profile.GetName())
	errors := deployGameserversForMatches(matches, backend)
	if errors != nil {
		log.Printf("Errors occurred while deploying game servers:")
		for _, e := range errors {
			log.Println(e.Error())
		}
	}
}

func main() {
	fmt.Println("Starting Director Service...")
	backend, err := NewBackend()
	if err != nil {
		log.Fatalf("Failed to create backend: %v", err)
	}
	for range time.Tick(time.Second * 5) {
		fmt.Println("Creating matches...")
		var wg sync.WaitGroup
		for _, profile := range BuildMatchProfiles() {
			wg.Add(1)
			go processMatches(&wg, profile, backend)
		}
		wg.Wait()
	}
}
