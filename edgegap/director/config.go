package main

import (
	"director/config"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"strings"
	"time"
)

type Matchmaker struct {
	Name      string
	UpdatedAt time.Time
	Config    *config.Matchmaker
}

func NewMatchmaker(name string, updatedAt string, config string) (*Matchmaker, error) {
	ua, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		log.Printf("NewMatchmaker::Failed to parse updatedAt input of %s", updatedAt)
		ua = time.Now()
	}
	mm := &Matchmaker{
		Name:      name,
		UpdatedAt: ua,
	}
	if config == "" {
		return nil, fmt.Errorf("NewMatchmaker::empty config")
	}
	if strings.HasPrefix(config, "{") {
		err = json.Unmarshal([]byte(config), &matchmaker.Config)
		if err != nil {
			return nil, fmt.Errorf("NewMatchmaker::Failed to unmarshal config as json!\nError: %s\nConfig: %s\n", err.Error(), config)
		}
	} else if strings.HasPrefix(config, "---") {
		err = yaml.Unmarshal([]byte(config), &matchmaker.Config)
		if err != nil {
			return nil, fmt.Errorf("NewMatchmaker::Failed to unmarshal config as yaml!\nError: %s\nConfig: %s\n", err.Error(), config)
		}
	} else {
		return matchmaker, fmt.Errorf("NewMatchmaker::Failed to unmarshal config. unknown format!\nConfig: %s\n", config)
	}
	return mm, nil
}
