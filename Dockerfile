FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /cf-auto-dns .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /cf-auto-dns /usr/local/bin/cf-auto-dns
ENTRYPOINT ["cf-auto-dns"]
