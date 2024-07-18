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

func NewResponse(ctx echo.Context, data interface{}) error {
	return ctx.JSON(http.StatusOK, Response{RequestId: ctx.Response().Header().Get(echo.HeaderXRequestID), Data: data})
}
func NewError(ctx echo.Context, code int) error {
	return ctx.JSON(code, Response{RequestId: ctx.Response().Header().Get(echo.HeaderXRequestID), Error: http.StatusText(code)})
}
