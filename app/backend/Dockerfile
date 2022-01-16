FROM golang:1.17

WORKDIR /go/src/app

COPY . .

RUN go get github.com/jmoiron/sqlx
RUN go get github.com/lib/pq
RUN go get github.com/ReneKroon/ttlcache/v2
RUN go get github.com/nats-io/nats.go
RUN go get github.com/nats-io/stan.go

RUN go build -o app cmd/my_project/main.go
CMD sleep 5 && ./app