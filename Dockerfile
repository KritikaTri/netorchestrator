FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api-gateway-full ./cmd/api-gateway-full

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/api-gateway-full /app/api-gateway-full
COPY config.yaml /app/config.yaml
EXPOSE 8080
ENTRYPOINT ["/app/api-gateway-full"]

