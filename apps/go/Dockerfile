FROM golang:1.22-alpine AS builder

# Install build tools, git, and C compiler for CGO
RUN apk add --no-cache git gcc musl-dev

# Create app directory
WORKDIR /app

# Copy go.mod and go.sum first, download dependencies (leverages Docker cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go binary with CGO enabled for dependencies that require it
RUN go build -o server ./main.go

# ----- Runtime image -----
FROM alpine:latest AS runtime

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server ./server

# Create uploads directory at runtime so that it can be mounted as a volume
RUN mkdir -p /app/uploads/audio

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup

# Change ownership of app directory to appuser
RUN chown -R appuser:appgroup /app

# The server listens on port 8080
EXPOSE 8080

# You can provide config via environment variables or mount an env file
# Example: docker run -d -p 8080:8080 --env ELEVENLABS_API_KEY=xxxxx open-reader-go
# or: docker run --env-file ./env -v $(pwd)/uploads:/app/uploads -p 8080:8080 open-reader-go

USER appuser

ENTRYPOINT ["/app/server"] 