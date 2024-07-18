package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Response struct {
	RequestId string      `json:"RequestId"`
	Data      interface{} `json:"Data,omitempty"`
	Error     string      `json:"Error,omitempty"`
}

type Context struct {
	echo.Context
}

func (c *Context) RespondError(code int) error {
	return c.JSON(code, Response{RequestId: c.Response().Header().Get(echo.HeaderXRequestID), Error: http.StatusText(code)})
}

func (c *Context) Respond(data interface{}) error {
	return c.JSON(http.StatusOK, Response{RequestId: c.Response().Header().Get(echo.HeaderXRequestID), Data: data})
}
