FROM golang:1.21 AS builder

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./app/cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

WORKDIR /config

COPY ./config/config.yaml .

WORKDIR /app

EXPOSE 8080

ENV TZ=Asia/Bangkok
ENV DOCKER_ENV=docker

CMD ["./main"]