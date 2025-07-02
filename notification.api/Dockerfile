FROM golang:1.24.3-alpine AS builder

WORKDIR /go/src/github.com/kenoboya/notification.api

COPY . .

RUN go mod download
RUN GOOS=linux go build -o ./api ./cmd/api/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /go/src/github.com/kenoboya/notification.api/api .

CMD ["./api"]
