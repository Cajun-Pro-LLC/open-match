package main

type OpenMatchError string

const (
	GRPCConnectionError OpenMatchError = "gRPC Connection Error"
)

func (e OpenMatchError) Error() string {
	return string(e)
}
