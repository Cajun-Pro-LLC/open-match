package main

import (
	swagger "github.com/cajun-pro-llc/edgegap-swagger"
	"log"
	"time"
)

type Matchmaker struct {
	Name      string
	UpdatedAt time.Time
	Config    *swagger.MatchmakerReleaseConfig
}

func NewMatchmaker(name string, updatedAt string, config *swagger.MatchmakerReleaseConfig) (*Matchmaker, error) {
	ua, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		log.Printf("NewMatchmaker::Failed to parse updatedAt input of %s", updatedAt)
		ua = time.Now()
	}
	mm := &Matchmaker{
		Name:      name,
		UpdatedAt: ua,
		Config:    config,
	}

	return mm, nil
}
