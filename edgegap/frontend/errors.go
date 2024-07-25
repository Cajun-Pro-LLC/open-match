package main

const (
	TicketNotFound      = RESTError("TICKET_NOT_FOUND")
	InvalidRequest      = RESTError("INVALID_REQUEST")
	CreateTicketFailure = RESTError("CREATE_TICKET_FAILURE")
)

type RESTError string

func (e RESTError) Error() string {
	return string(e)
}
