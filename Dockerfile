FROM golang:1.23.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o leader-elector .

FROM alpine:latest
COPY --from=builder /app/leader-elector /leader-elector
CMD ["/leader-elector"]
