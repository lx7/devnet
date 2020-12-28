FROM golang:1.15-buster

ENV GO111MODULE=on

WORKDIR /go/src/devnet

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /go/src/devnet/cmd/signald
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /signald

FROM alpine:3.12.3

RUN apk --no-cache add ca-certificates
COPY --from=0 /signald /usr/local/bin/signald

COPY configs/docker/signald.yaml /etc/devnet/signald.yaml

CMD ["/usr/local/bin/signald"]

