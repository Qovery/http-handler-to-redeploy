# Build stage
FROM golang:1.19-alpine as builder

# Set up necessary build tools
RUN apk update && apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o main main.go

# Final stage
FROM alpine:3.15

# Set the working directory
WORKDIR /app

# Install necessary runtime dependencies
RUN apk update && apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/main /app/main

# Expose the application port
EXPOSE 8000

# Run the binary
CMD ["/app/main"]
