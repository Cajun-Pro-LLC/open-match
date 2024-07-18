package main

type OpenMatchError string

const (
	GRPCConnectionError OpenMatchError = "gRPC Connection Error"
)
