package main

import (
	"fmt"
	swagger "github.com/cajun-pro-llc/edgegap-swagger"
	"open-match.dev/open-match/pkg/pb"
	"strings"
)

type Gameserver struct {
	Players    []*playerDetails
	Connection string
	match      *pb.Match
}

func newGameServer(m *pb.Match) *Gameserver {
	details := &Gameserver{}

	for _, t := range m.GetTickets() {
		details.Players = append(details.Players, newPlayerDetails(t))
	}
	return details
}

func (gs *Gameserver) GetPlayerIps() []string {
	IPs := make([]string, len(gs.Players))
	for i, p := range gs.Players {
		IPs[i] = p.IP
	}
	return IPs
}

func (gs *Gameserver) GetPlayerIds() []string {
	IDs := make([]string, len(gs.Players))
	for i, p := range gs.Players {
		IDs[i] = p.PlayerId
	}
	return IDs
}

func (gs *Gameserver) GetTicketIds() []string {
	TicketIds := make([]string, len(gs.Players))
	for i, t := range gs.Players {
		TicketIds[i] = t.TicketId
	}
	return TicketIds
}

func (gs *Gameserver) getDeployModel() swagger.DeployModel {
	envVars := []swagger.DeployEnvModel{
		{Key: "PlayerIds", Value: strings.Join(gs.GetPlayerIds(), ",")},
	}
	matchProfile := gs.getMatchProfile()
	for _, selector := range matchProfile.Selectors {
		if selector.InjectEnv {
			envVar := swagger.DeployEnvModel{
				Key:   selector.Key,
				Value: string(gs.match.Extensions[selector.Key].GetValue()),
			}
			envVars = append(envVars, envVar)
		}

	}
	return swagger.DeployModel{
		AppName:     matchProfile.App,
		VersionName: matchProfile.Version,
		IpList:      gs.GetPlayerIps(),
		EnvVars:     envVars,
	}
}

func (gs *Gameserver) getMatchProfile() *MatchmakerProfile {
	for _, profile := range matchmaker.Config.Profiles {
		if "profile_"+profile.Id == gs.match.GetMatchProfile() {
			return profile
		}
	}
	fmt.Printf("Could not find match profile for %s. Using first profile\n", gs.match.GetMatchProfile())
	return matchmaker.Config.Profiles[0]
}
