FROM golang:alpine AS go
ARG FUNCTION_NAME
WORKDIR /app
ENV GO111MODULE=on
COPY /edgegap/$FUNCTION_NAME .
RUN go mod tidy && go build -o main .
CMD ["/app/main"]