FROM golang:1.16-alpine AS builder

RUN apk update && apk add git ca-certificates tzdata gcc g++

COPY cmd /cmd
COPY config /config
COPY services /services
COPY go.mod /go.mod
COPY go.sum /go.sum
COPY main.go /main.go

WORKDIR /

RUN go mod tidy

RUN GOOS=linux go build -a -installsuffix nocgo -o /minter-sentinel .

FROM alpine

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo/
COPY --from=builder /minter-sentinel /usr/bin/minter-sentinel

ENTRYPOINT ["/usr/bin/minter-sentinel"]
CMD ["--config", "/config.yaml", "start"]