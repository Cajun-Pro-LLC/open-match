FROM golang:1.22-alpine AS build
ARG FUNCTION_NAME
WORKDIR /app
ENV GO111MODULE=on
COPY . .
WORKDIR /app/edgegap/${FUNCTION_NAME}
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /app/main ./...

FROM gcr.io/distroless/static-debian11
COPY --from=build /app/main /app

ENTRYPOINT ["/app"]