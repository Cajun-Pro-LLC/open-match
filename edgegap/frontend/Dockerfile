FROM golang:alpine as go
WORKDIR /app
ENV GO111MODULE=on

COPY . .
RUN go mod tidy && \
    go build -o frontend .

CMD ["/app/frontend"]