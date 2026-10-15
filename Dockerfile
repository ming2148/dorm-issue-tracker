# Multi-stage build for Go Application
FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o dorm-tracker cmd/api/main.go

# Production minimal scratch/alpine runner
FROM alpine:3.19

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/dorm-tracker .
RUN mkdir -p /app/uploads /app/data

EXPOSE 8080

CMD ["./dorm-tracker"]
