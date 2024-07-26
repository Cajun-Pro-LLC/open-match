package main

import (
	swagger "github.com/cajun-pro-llc/edgegap-swagger"
	"open-match.dev/open-match/pkg/pb"
	"strings"
)

type Gameserver struct {
	players    GameserverPlayers
	Connection string
	match      *pb.Match
}

type GameserverPlayers []*playerDetails

func newGameServer(m *pb.Match) *Gameserver {
	details := &Gameserver{
		players: []*playerDetails{},
		match:   m,
	}

	for _, t := range m.GetTickets() {
		details.players = append(details.players, newPlayerDetails(t))
	}
	return details
}

func (gs GameserverPlayers) IPs() []string {
	IPs := make([]string, len(gs))
	for i, p := range gs {
		IPs[i] = p.IP
	}
	return IPs
}

func (gs GameserverPlayers) IDs() []string {
	IDs := make([]string, len(gs))
	for i, p := range gs {
		IDs[i] = p.PlayerId
	}
	return IDs
}

func (gs GameserverPlayers) Tickets() []string {
	TicketIds := make([]string, len(gs))
	for i, t := range gs {
		TicketIds[i] = t.TicketId
	}
	return TicketIds
}

func (gs *Gameserver) Players() GameserverPlayers {
	return gs.players
}

func (gs *Gameserver) DeployModel() swagger.DeployModel {
	matchProfile := gs.getMatchProfile()
	envVars := []swagger.DeployEnvModel{
		{Key: "MATCH_PLAYER_IDS", Value: strings.Join(gs.Players().IDs(), ",")},
		{Key: "MATCH_PROFILE", Value: matchProfile.Id},
	}

	for _, selector := range matchProfile.Selectors {
		if selector.InjectEnv {
			envVar := swagger.DeployEnvModel{
				Key:   "MATCH_" + strings.ToUpper(selector.Key),
				Value: gs.match.Tickets[0].SearchFields.StringArgs[selector.Key],
			}
			envVars = append(envVars, envVar)
		}

	}
	return swagger.DeployModel{
		AppName:     matchProfile.App,
		VersionName: matchProfile.Version,
		IpList:      gs.Players().IPs(),
		EnvVars:     envVars,
	}
}

func (gs *Gameserver) getMatchProfile() *swagger.MatchmakerProfile {
	for _, profile := range matchmaker.Config.Profiles {
		if "profile_"+profile.Id == gs.match.GetMatchProfile() {
			return profile
		}
	}
	log.Error().Str("matchProfile", gs.match.GetMatchProfile()).Msg("Could not find matching match profile config. Using first profile")
	return matchmaker.Config.Profiles[0]
}

func (gs *Gameserver) GamePort() string {
	return gs.getMatchProfile().GamePort
}
