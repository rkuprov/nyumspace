FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./applications/nyumspace/api

# Runtime stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /api .

EXPOSE 8080

ENTRYPOINT ["./api"]
