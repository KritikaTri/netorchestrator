# syntax=docker/dockerfile:1
# One Dockerfile for every service:  docker build -f deploy/Dockerfile --build-arg SERVICE=gateway -t netorch/gateway .
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SERVICE
RUN test -n "$SERVICE" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/svc ./cmd/${SERVICE}

# distroless: no shell, no package manager, runs as a non-root user
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/svc /svc
USER nonroot:nonroot
ENTRYPOINT ["/svc"]
