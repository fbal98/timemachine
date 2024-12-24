FROM golang:1.21.5-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o timemachine main.go

# Create final lightweight image
FROM alpine:latest

WORKDIR /app

# Install timezone data
RUN apk add --no-cache tzdata

# Copy the binary from builder
COPY --from=builder /app/timemachine .
COPY --from=builder /app/.env .
COPY --from=builder /app/messages.json .

# Set timezone
ENV TZ=Asia/Muscat

# Expose HTTP port
EXPOSE 8080

# Run the application
CMD ["./timemachine"] 