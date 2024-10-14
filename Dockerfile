FROM golang:1.21.0 AS builder
LABEL authors="vfaergestad"

WORKDIR /app

# Set environment variables for static linking

COPY cmd cmd
COPY config config
COPY logger logger
COPY models models
COPY utils utils
COPY websocket websocket
COPY filewriter filewriter
COPY go.mod .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o mattermost-logger cmd/main.go

FROM scratch AS runtime

WORKDIR /app

COPY --from=builder /app/mattermost-logger .

ENTRYPOINT ["./mattermost-logger"]
