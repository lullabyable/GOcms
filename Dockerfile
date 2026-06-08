FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o gocms ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/gocms .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config ./config
EXPOSE 8080
CMD ["./gocms"]
