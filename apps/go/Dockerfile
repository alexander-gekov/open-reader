FROM golang:1.22-alpine AS builder

# Install build tools and git for go modules that require it
RUN apk add --no-cache git

# Create app directory
WORKDIR /app

# Copy go.mod and go.sum first, download dependencies (leverages Docker cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go binary (CGO disabled for a fully static binary)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./main.go

# ----- Runtime image -----
FROM gcr.io/distroless/static-debian11 AS runtime

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server ./server

# Create uploads directory at runtime so that it can be mounted as a volume
RUN mkdir -p /app/uploads/audio

# The server listens on port 8080
EXPOSE 8080

# You can provide config via environment variables or mount an env file
# Example: docker run -d -p 8080:8080 --env ELEVENLABS_API_KEY=xxxxx open-reader-go
# or: docker run --env-file ./env -v $(pwd)/uploads:/app/uploads -p 8080:8080 open-reader-go

USER nonroot

ENTRYPOINT ["/app/server"] 