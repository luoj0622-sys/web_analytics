FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

ARG CMD
RUN go build -o /app/service ./cmd/${CMD}


FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

ARG CMD
COPY --from=builder /app/service /app/service

ENTRYPOINT ["/app/service"]
