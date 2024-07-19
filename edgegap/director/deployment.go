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
	client *swagger.APIClient
	ctx    context.Context
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

func (a *Arbitrum) sendGameserverDeployment(deployment swagger.DeployModel) (*swagger.Request, error) {
	request, _, err := a.client.DeploymentsApi.Deploy(a.ctx, deployment)
	if err != nil {
		log.Printf("Could not deploy game server, err: %v", err.Error())
		return nil, err
	}
	return &request, nil
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
	request, err := a.sendGameserverDeployment(gs.getDeployModel())
	if err != nil {
		return err
	}

	// Wait for server ready
	response, err := a.waitForGameServerReady(request)
	if err != nil {
		return err
	}

	gs.Connection = fmt.Sprintf("%s:%d", response.PublicIp, response.Ports[gs.GamePort].External)
	return nil
}
