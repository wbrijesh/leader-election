# Use the official Golang image as the base image
FROM golang:1.23.3-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go modules and main.go file into the container
COPY go.mod ./
COPY main.go ./

# Download the Go dependencies
RUN go mod tidy

# Build the Go program
RUN go build -o server main.go

# Start a new, smaller image from Alpine
FROM alpine:latest

# Copy the compiled Go program from the builder stage
COPY --from=builder /app/server /server

# Expose port 8080
EXPOSE 8080

# Define the command to run the Go server
CMD ["/server"]
