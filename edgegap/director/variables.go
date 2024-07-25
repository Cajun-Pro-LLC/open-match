package main

import (
	"github.com/cajun-pro-llc/open-match/utils"
)

var (
	matchmaker *Matchmaker
	log        = utils.NewLogger(map[string]string{"service": "director"})
)
