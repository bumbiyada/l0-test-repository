FROM golang:1.17

WORKDIR /go/src/app

COPY . .

RUN go get github.com/nats-io/nats.go
RUN go get github.com/nats-io/stan.go

RUN go build -o app main.go

CMD sleep 4 && ./app -s nats://nats-server:4222 foo {"id": "1", "val": "test"}
# CMD ["./app", "foo", "test"]