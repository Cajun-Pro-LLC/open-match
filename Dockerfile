ARG FUNCTION_NAME
FROM golang:alpine as go
WORKDIR /app
ENV GO111MODULE=on
COPY /edgegap/${FUNCTION_NAME} .
RUN go mod tidy && go build -o main .
CMD ["/app/main"]