# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency definitions first for efficient caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the project source code
COPY . .

# Build the binary targeting the main file inside src/
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./src/main.go

# Production Stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]