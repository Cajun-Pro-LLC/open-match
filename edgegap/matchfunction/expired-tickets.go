package main

import (
	"fmt"
	"net/http"
	"open-match.dev/open-match/pkg/pb"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func findExpiredTickets(poolTickets map[string][]*pb.Ticket) []*pb.Ticket {
	var expiredTickets []*pb.Ticket
	for pool, tickets := range poolTickets {
		var validPoolTickets []*pb.Ticket
		for _, ticket := range tickets {
			if time.Now().After(ticket.GetCreateTime().AsTime().Add(maxTicketTTL * time.Second)) {
				expiredTickets = append(expiredTickets, ticket)
			} else {
				validPoolTickets = append(validPoolTickets, ticket)
			}
		}
		poolTickets[pool] = validPoolTickets
	}
	return expiredTickets
}

func deleteTickets(tickets []*pb.Ticket) {
	var wg sync.WaitGroup
	for _, ticket := range tickets {
		wg.Add(1)
		go func(ticket *pb.Ticket) {
			defer wg.Done()
			err := deleteTicket(ticket.GetId())
			if err != nil {
				log.Err(err).Msg("unable to delete ticket")
			}
		}(ticket)
	}

	wg.Wait()
}

func deleteTicket(ticketId string) error {
	r := regexp.MustCompile(`-(custom-frontend|mmf|director)-[a-z0-9]+-[a-z0-9]+$`)
	prefix := r.ReplaceAllString(os.Getenv("HOSTNAME"), "")
	underlined := strings.ReplaceAll(prefix, "-", "_")
	upper := strings.ToUpper(underlined)
	host := os.Getenv(fmt.Sprintf("%s_CUSTOM_FRONTEND_SVC_SERVICE_HOST", upper))
	//goland:noinspection HttpUrlsUsage
	url := fmt.Sprintf("http://%s:51504/v1/tickets/%s", host, ticketId)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}
