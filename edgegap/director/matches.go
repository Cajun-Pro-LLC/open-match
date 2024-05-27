package main

import (
	"fmt"
	"open-match.dev/open-match/pkg/pb"
)

type Mode string

const (
	Dev      Mode = "dev"
	Unranked Mode = "unranked"
	Ranked   Mode = "ranked"
	Event    Mode = "event"
)

func (c Mode) String() string {
	return string(c)
}

type Category string

const (
	OneVsOne   Category = "1v1"
	FreeForAll Category = "ffa"
	Teams      Category = "teams"
	Royale     Category = "royale"
)

func (c Category) String() string {
	return string(c)
}

func buildMatchProfile(mode Mode, category Category) *pb.MatchProfile {
	modeStr := mode.String()
	categoryStr := category.String()
	profileName := fmt.Sprintf("%s_%s_profile", modeStr, categoryStr)
	poolName := fmt.Sprintf("pool_%s_%s", modeStr, categoryStr)

	return &pb.MatchProfile{
		Name: profileName,
		Pools: []*pb.Pool{
			{
				Name: poolName,
				TagPresentFilters: []*pb.TagPresentFilter{
					{Tag: modeStr},
					{Tag: categoryStr},
				},
			},
		},
	}
}

func BuildMatchProfiles() []*pb.MatchProfile {
	var profiles []*pb.MatchProfile
	modes := []Mode{Dev, Unranked, Ranked, Event}
	categories := []Category{OneVsOne, FreeForAll, Teams, Royale}
	for _, mode := range modes {
		for _, category := range categories {
			profile := buildMatchProfile(mode, category)
			profiles = append(profiles, profile)
		}
	}
	return profiles
}
