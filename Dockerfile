# syntax=docker/dockerfile:1

# Fetch
FROM golang:latest AS fetch-stage
COPY go.mod go.sum /app
WORKDIR /app
RUN go mod download

# Generate
FROM ghcr.io/a-h/templ:latest AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

# Build
FROM golang:1.22 AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app

COPY . .
RUN go build -o=./bin/main ./cmd

EXPOSE 9001

CMD ["./bin/main"]
