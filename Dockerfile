# Multi-stage build for Go Application
FROM golang:alpine AS builder

# Enable toolchain auto-switching if needed
ENV GOTOOLCHAIN=auto

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o dorm-tracker cmd/api/main.go

# Production minimal scratch/alpine runner
FROM alpine:latest

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/dorm-tracker .
RUN mkdir -p /app/uploads /app/data

EXPOSE 8080

CMD ["./dorm-tracker"]
