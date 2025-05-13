FROM golang:1.23.0-alpine AS builder

WORKDIR /sentinel

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/

FROM alpine:latest

WORKDIR /sentinel

COPY --from=builder /sentinel/main .

RUN mkdir -p /var/log/sentinel

CMD ["./main"]

