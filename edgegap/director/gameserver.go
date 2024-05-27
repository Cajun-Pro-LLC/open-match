package main

import (
	swagger "github.com/cajun-pro-llc/edgegap-swagger"
	"open-match.dev/open-match/pkg/pb"
	"strings"
)

type Gameserver struct {
	GamePort   string
	GameMode   Mode
	Category   Category
	Players    []*playerDetails
	MatchId    string
	Connection string
}

func newGameServer(m *pb.Match) *Gameserver {
	details := &Gameserver{
		MatchId:  m.GetMatchId(),
		GamePort: gameServerPort,
		GameMode: Mode(m.Extensions["mode"].Value),
		Category: Category(m.Extensions["category"].Value),
	}

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
	return swagger.DeployModel{
		AppName:     appName,
		VersionName: appVersion,
		IpList:      gs.GetPlayerIps(),
		EnvVars: []swagger.DeployEnvModel{
			{Key: "Mode", Value: gs.GameMode.String()},
			{Key: "Category", Value: gs.Category.String()},
			{Key: "PlayerIds", Value: strings.Join(gs.GetPlayerIds(), ",")},
		},
	}
}
