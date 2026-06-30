FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o hobby-api ./cmd/app
RUN go build -o migrate ./cmd/migrate

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/hobby-api .
COPY --from=builder /app/migrate .
COPY config.yml .

EXPOSE 8080

CMD ["./hobby-api"]