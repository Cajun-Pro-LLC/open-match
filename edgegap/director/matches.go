package main

import (
	"fmt"
	swagger "github.com/cajun-pro-llc/edgegap-swagger"
	"github.com/cajun-pro-llc/open-match/utils"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"open-match.dev/open-match/pkg/pb"
)

func buildPools(profile *swagger.MatchmakerProfile, matchProfiles *[]*pb.MatchProfile, tempProfile []*pb.Pool, index int) {
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
			Name:               fmt.Sprintf("pool_%s_%s", profile.Id, item),
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
		utils.PrettyPrint("Generating pool "+tempProfile[index].Name, tempProfile[index])
		buildPools(profile, matchProfiles, tempProfile, index+1)
	}
}

func buildMatchmakerProfile(profile *swagger.MatchmakerProfile) []*pb.MatchProfile {
	var matchProfiles []*pb.MatchProfile
	matchProfile := &pb.MatchProfile{
		Name:  "profile_" + profile.Id,
		Pools: []*pb.Pool{},
		Extensions: map[string]*anypb.Any{
			"playerCount": {
				TypeUrl: "type.googleapis.com/google.protobufInt32Value",
				Value:   utils.Int32ToProto(profile.MatchPlayerCount),
			},
		},
	}
	utils.PrettyPrint("Generating profile "+matchProfile.Name, matchProfile)
	matchProfiles = append(matchProfiles, matchProfile)

	// Calling our recursive function
	buildPools(profile, &matchProfiles, make([]*pb.Pool, len(profile.Selectors)), 0)

	return matchProfiles
}

func buildMatchmakerProfiles(profiles []*swagger.MatchmakerProfile) []*pb.MatchProfile {
	var matchProfiles []*pb.MatchProfile
	for i := range profiles {
		matchProfiles = append(matchProfiles, buildMatchmakerProfile(profiles[i])...)
	}
	return matchProfiles
}
