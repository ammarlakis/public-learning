FROM golang:1.24 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

RUN ls -l /

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/server /app/server
RUN ls -l
EXPOSE 8080

CMD ["/app/server"]