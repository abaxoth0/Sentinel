# Build stage
FROM golang:1.23.0-alpine AS builder

WORKDIR /sentinel

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/

# Copy config files to root
RUN cp cmd/.env cmd/*.config.* cmd/RBAC.json ./

# Final stage
FROM alpine:latest

WORKDIR /sentinel

# Copy files from builder
COPY --from=builder /sentinel/main .
COPY --from=builder /sentinel/.env .
COPY --from=builder /sentinel/*.config.* ./
COPY --from=builder /sentinel/RBAC.json ./

# Command to run the executable
CMD ["./main"]

