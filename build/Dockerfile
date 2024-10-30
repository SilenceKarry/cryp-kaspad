FROM golang:1.23 AS build

ENV GO111MODULE=on

WORKDIR /app

COPY . .

RUN go mod download

RUN GOOS=linux GOARCH=amd64 go build -o cryp-kaspad-server ./cmd/cryp-kaspad/main.go

FROM oraclelinux:9-slim

WORKDIR /app

COPY --from=build /app/cryp-kaspad-server /app/cryp-kaspad-server
COPY --from=build /app/configs/config_dev.yaml /app/config.yaml

CMD ["/app/cryp-kaspad-server"]
