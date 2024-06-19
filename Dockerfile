FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/bin/main .
COPY .env .

EXPOSE 8080
