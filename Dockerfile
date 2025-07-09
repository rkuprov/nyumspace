# Use the pre-built base image
FROM nyumspace-base:latest

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o ./bin/serves ./cmd/serves

# Set default command
CMD ["./bin/serves"]
