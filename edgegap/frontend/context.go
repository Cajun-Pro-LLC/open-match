package main

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"net/http"
)

// CreateTicketRequest represents the model that we should receive for our create ticket endpoint
type CreateTicketRequest struct {
	ProfileId   string         `json:"edgegap_profile_id"`
	PlayerId    string         `json:"player_id"`
	Matchmaking CTRMatchmaking `json:"matchmaking_data"`
}

type CTRMatchmaking struct {
	Selectors CTRMatchmakingSelectors `json:"selector_data"`
	Filters   CTRMatchmakingFilters   `json:"filter_data"`
}

type CTRMatchmakingSelectors map[string]string
type CTRMatchmakingFilters map[string]float64

func (c CreateTicketRequest) MarshalZerologObject(e *zerolog.Event) {
	e.Str("edgegap_profile_id", c.ProfileId)
	e.Str("player_id", c.PlayerId)
	e.Object("matchmaking_data", c.Matchmaking)
}

func (c CTRMatchmaking) MarshalZerologObject(e *zerolog.Event) {
	e.Object("selector_data", c.Selectors)
	e.Object("filter_data", c.Filters)
}

func (c CTRMatchmakingSelectors) MarshalZerologObject(e *zerolog.Event) {
	for k, v := range c {
		e.Str(k, v)
	}
}

func (c CTRMatchmakingFilters) MarshalZerologObject(e *zerolog.Event) {
	for k, v := range c {
		e.Float64("filter_"+k, v)
	}
}

type Response struct {
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

type Context struct {
	echo.Context
}

func (c *Context) RespondErrorCustom(code int, err interface{}) error {
	e := "unknown error type"
	switch err.(type) {
	case string:
		e = err.(string)
	case error:
		e = err.(error).Error()
	}
	return c.JSON(code, Response{RequestId: c.Response().Header().Get(echo.HeaderXRequestID), Error: e})
}

func (c *Context) RespondError(code int) error {
	return c.JSON(code, Response{RequestId: c.Response().Header().Get(echo.HeaderXRequestID), Error: http.StatusText(code)})
}

func (c *Context) Respond(data interface{}) error {
	return c.JSON(http.StatusOK, Response{RequestId: c.Response().Header().Get(echo.HeaderXRequestID), Data: data})
}
