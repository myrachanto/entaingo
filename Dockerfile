# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum for dependency download
COPY go.mod .
COPY go.sum .

# Download the dependencies
RUN go mod download

# Copy the remaining application code
COPY . .

# Build the application
RUN go build -o entaingo .

# Run stage
FROM alpine

WORKDIR /app

# Copy the binary from the builder stage to the run stage
COPY --from=builder /app/entaingo .

# Copy environment configuration files if needed
COPY app.env .
COPY .env .

# Expose the application port
EXPOSE 4000

# Start the application
CMD ["/app/entaingo"]
