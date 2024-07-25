package main

import (
	"context"
	"errors"
	"fmt"
	swagger "github.com/cajun-pro-llc/edgegap-swagger"
	"os"
	"strconv"
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
	apiKey := swagger.APIKey{
		Key:    os.Getenv("ARBITRUM_API_KEY"),
		Prefix: "token",
	}
	return &Arbitrum{
		client: swagger.NewAPIClient(configuration),
		ctx:    context.WithValue(context.Background(), swagger.ContextAPIKey, apiKey),
	}
}

func (a *Arbitrum) waitForGameServerReady(request *swagger.Request) (*swagger.Status, error) {
	timeout := 60.0
	envTimeout := os.Getenv("DEPLOY_TIMEOUT")
	if envTimeout != "" {
		t, err := strconv.Atoi(envTimeout)
		if err == nil && t > 10 {
			timeout = float64(t)
		}
	}
	start := time.Now()
	status := ""
	var err error
	var response swagger.Status
	// Waiting for the server to be ready
	for status != "Status.READY" && time.Since(start).Seconds() <= timeout {
		response, _, err = a.client.DeploymentsApi.DeploymentStatusGet(a.ctx, request.RequestId)
		if err != nil {
			log.Err(err).Msg("error fetching status")
			continue
		}
		log.Trace().Str("function", "waitForGameServerReady").Str("requestId", request.RequestId).Str("status", response.CurrentStatus).Msg("got deployment status")
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
	request, _, err := a.client.DeploymentsApi.Deploy(a.ctx, gs.DeployModel())
	if err != nil {
		log.Err(err).Msg("Could not deploy game server")
		return err
	}

	// Wait for server ready
	response, err := a.waitForGameServerReady(&request)
	if err != nil {
		return err
	}

	gs.Connection = fmt.Sprintf("%s:%d", response.PublicIp, response.Ports[gs.GamePort()].External)
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
	fmt.Printf("Loading configuration: %v\n", resp.Configuration)
	a.matchmaker = NewMatchmaker(resp.Name, resp.LastUpdated, &resp.Configuration)
	return nil
}
