# Build Stage - Use latest stable Go image or golang:1.26-alpine
FROM golang:alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary
RUN GOOS=linux go build -o server ./src/main.go

# Production Stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]