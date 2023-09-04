# Stage 1: Build the Go application
FROM golang:1.21 AS builder
LABEL authors="the-eduardo"

WORKDIR /app
COPY main.go go.mod go.sum ./
COPY repos.json ./
RUN CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o app

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Install CA certificates
RUN apk --no-cache add ca-certificates

# Copy the built executable from the "builder" stage
COPY --from=builder /app/app /app
COPY --from=builder /app/repos.json /repos.json

ENTRYPOINT ["/app"]