FROM golang:1.18-alpine as builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server

# Use a minimal alpine image for the final container
FROM alpine:latest

# Add CA certificates
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server /app/server

# Copy static and template files
COPY static/ /app/static/
COPY templates/ /app/templates/

# Make sure the container can write to users.json, chats.json, and recentChats.json
COPY users.json /app/users.json
COPY chats.json /app/chats.json
COPY recentChats.json /app/recentChats.json

# Expose the port the app runs on
EXPOSE 8080

# Run the application
CMD ["/app/server"] 