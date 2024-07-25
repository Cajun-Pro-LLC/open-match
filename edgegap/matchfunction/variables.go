package main

import "github.com/cajun-pro-llc/open-match/utils"

const (
	matchName    = "match-function"
	maxTicketTTL = 600
)

var log = utils.NewLogger(map[string]string{"service": "matchfunction"})
