FROM golang:1.24-alpine AS builder

WORKDIR /go/src/github.com/kenoboya/profile.api

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /go/src/github.com/kenoboya/profile.api/api .

EXPOSE 8081
EXPOSE 50053

CMD ["./api"]
