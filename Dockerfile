# syntax=docker/dockerfile:1

# -------- Builder --------
FROM golang:1.24 AS builder
WORKDIR /app

# Set Go build environment for native architecture and static linking
ENV GOOS=linux
ENV CGO_ENABLED=0

# Detect native architecture at build time
ARG TARGETARCH
ENV GOARCH=${TARGETARCH}

# Enable Go module caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the API gateway binary (native architecture, static linking)
# Use reduced parallelism to avoid OOM during compilation
RUN GOMAXPROCS=2 go build -ldflags="-w -s" -trimpath -o /app/bin/api-gateway ./cmd/api-gateway

# -------- Runtime --------
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

# Copy binary and minimal runtime assets
COPY --from=builder /app/bin/api-gateway /app/api-gateway
COPY --from=builder /app/config.yaml /app/config.yaml

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/api-gateway"]
