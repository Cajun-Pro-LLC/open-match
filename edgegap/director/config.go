package main

import (
	swagger "github.com/cajun-pro-llc/edgegap-swagger"
	"strings"
	"time"
)

type Matchmaker struct {
	Name      string
	UpdatedAt time.Time
	Config    *swagger.MatchmakerReleaseConfig
}

func NewMatchmaker(name string, updatedAt string, config *swagger.MatchmakerReleaseConfig) *Matchmaker {
	mm := &Matchmaker{
		Name:   name,
		Config: config,
	}
	ua, err := time.Parse(time.RFC3339, strings.Replace(updatedAt, " ", "T", 1))
	if err != nil {
		log.Err(err).Str("function", "NewMatchmaker").Str("input", updatedAt).Msg("Failed to parse updatedAt")
		mm.UpdatedAt = time.Now()
	} else {
		mm.UpdatedAt = ua
	}
	return mm
}
