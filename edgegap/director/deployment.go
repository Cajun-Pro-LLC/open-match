package main

import (
	"context"
	"errors"
	"fmt"
	swagger "github.com/cajun-pro-llc/edgegap-swagger"
	"log"
	"os"
	"time"
)

type Arbitrum struct {
	client     *swagger.APIClient
	ctx        context.Context
	matchmaker *Matchmaker
}

func newArbitrum() *Arbitrum {
	// Creating API Client to communicate with arbitrium
	configuration := swagger.NewConfiguration()
	configuration.BasePath = os.Getenv("ARBITRIUM_API_URL")
	return &Arbitrum{
		client: swagger.NewAPIClient(configuration),
		ctx: context.WithValue(context.Background(), swagger.ContextAPIKey, swagger.APIKey{
			Key:    os.Getenv("ARBITRUM_API_KEY"),
			Prefix: "token",
		}),
	}
}

func (a *Arbitrum) waitForGameServerReady(request *swagger.Request) (*swagger.Status, error) {
	timeout := 30.0
	start := time.Now()
	status := ""
	var response swagger.Status
	// Waiting for the server to be ready
	for status != "Status.READY" && time.Since(start).Seconds() <= timeout {
		response, _, err := a.client.DeploymentsApi.DeploymentStatusGet(a.ctx, request.RequestId)
		if err != nil {
			log.Printf("Error while fetching status, err: %v", err.Error())
			continue
		}
		status = response.CurrentStatus
		time.Sleep(1 * time.Second) //	let's wait a bit
	}
	if time.Since(start).Seconds() > timeout {
		return nil, errors.New("timeout while waiting for deployment")
	}
	return &response, nil
}

func (a *Arbitrum) DeployGameserver(gs *Gameserver) error {
	// Perform deployment
	request, _, err := a.client.DeploymentsApi.Deploy(a.ctx, gs.getDeployModel())
	if err != nil {
		log.Printf("Could not deploy game server, err: %v", err.Error())
		return err
	}

	// Wait for server ready
	response, err := a.waitForGameServerReady(&request)
	if err != nil {
		return err
	}

	gs.Connection = fmt.Sprintf("%s:%d", response.PublicIp, response.Ports[gs.getMatchProfile().GamePort].External)
	return nil
}

func (a *Arbitrum) LoadConfiguration() error {
	configName := os.Getenv("CONFIG_NAME")
	if configName == "" {
		configName = "default"
	}
	resp, _, err := a.client.MatchmakerApi.GetMatchmakerReleaseConfig(a.ctx, configName)
	if err != nil {
		log.Printf("Error laoding configuration: %s", err.Error())
		return err
	}
	a.matchmaker, err = NewMatchmaker(resp.Name, resp.UpdatedAt, resp.Configuration)
	if err != nil {
		return err
	}
	return nil
}
