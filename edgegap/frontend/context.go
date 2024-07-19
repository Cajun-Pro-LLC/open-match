package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Response struct {
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

type Context struct {
	echo.Context
}

func (c *Context) RespondErrorCustom(code int, err string) error {
	return c.JSON(code, Response{RequestId: c.Response().Header().Get(echo.HeaderXRequestID), Error: err})
}

func (c *Context) RespondError(code int) error {
	return c.JSON(code, Response{RequestId: c.Response().Header().Get(echo.HeaderXRequestID), Error: http.StatusText(code)})
}

func (c *Context) Respond(data interface{}) error {
	return c.JSON(http.StatusOK, Response{RequestId: c.Response().Header().Get(echo.HeaderXRequestID), Data: data})
}
