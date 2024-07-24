package main

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"open-match.dev/open-match/pkg/pb"
)

func buildPools(profile *MatchmakerProfile, matchProfiles *[]*pb.MatchProfile, tempProfile []*pb.Pool, index int) {
	if index == len(profile.Selectors) {
		pools := make([]*pb.Pool, len(tempProfile))
		copy(pools, tempProfile)

		// Append pools to the last created profile
		if len(*matchProfiles) > 0 {
			(*matchProfiles)[len(*matchProfiles)-1].Pools = append((*matchProfiles)[len(*matchProfiles)-1].Pools, pools...)
		}

		return
	}
	var filters []*pb.DoubleRangeFilter
	for _, filter := range profile.Filters {
		f := &pb.DoubleRangeFilter{
			DoubleArg: filter.Name,
			Max:       filter.Maximum,
			Min:       filter.Minimum,
			Exclude:   pb.DoubleRangeFilter_NONE,
		}
		filters = append(filters, f)
	}
	for _, item := range profile.Selectors[index].Items {
		tempProfile[index] = &pb.Pool{
			Name:               fmt.Sprintf("pool_%s_%s", profile.Selectors[index].Name, item),
			DoubleRangeFilters: filters,
			StringEqualsFilters: []*pb.StringEqualsFilter{
				{
					StringArg: profile.Selectors[index].Name,
					Value:     item,
				},
			},
			TagPresentFilters: []*pb.TagPresentFilter{
				{Tag: profile.Id},
			},
			CreatedAfter: timestamppb.New(matchmaker.UpdatedAt),
		}
		buildPools(profile, matchProfiles, tempProfile, index+1)
	}
}

func buildMatchmakerProfile(profile *MatchmakerProfile) []*pb.MatchProfile {
	var matchProfiles []*pb.MatchProfile
	playerCountPb := &wrapperspb.Int32Value{Value: profile.MatchPlayerCount}
	playerCount, err := proto.Marshal(playerCountPb)
	if err != nil {
		fmt.Printf("Error marshaling PlayerCount: %s\n", err.Error())
	}
	matchProfile := &pb.MatchProfile{
		Name:  "profile_" + profile.Id,
		Pools: []*pb.Pool{},
		Extensions: map[string]*anypb.Any{
			"playerCount": {
				TypeUrl: "type.googleapis.com/google.protobufInt32Value",
				Value:   playerCount,
			},
		},
	}
	matchProfiles = append(matchProfiles, matchProfile)

	// Calling our recursive function
	buildPools(profile, &matchProfiles, make([]*pb.Pool, len(profile.Selectors)), 0)

	return matchProfiles
}

func buildMatchmakerProfiles(profiles []*MatchmakerProfile) []*pb.MatchProfile {
	var matchProfiles []*pb.MatchProfile
	for i := range profiles {
		matchProfiles = append(matchProfiles, buildMatchmakerProfile(profiles[i])...)
	}
	return matchProfiles
}
