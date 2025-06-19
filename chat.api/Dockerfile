FROM golang:1.24.3-alpine AS builder

WORKDIR /go/src/github.com/kenoboya/chat.api

COPY . .

RUN go mod download
RUN GOOS=linux go build -o ./api ./cmd/api/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /go/src/github.com/kenoboya/chat.api/api .
COPY .env .

COPY configs /configs

CMD ["./api"]
