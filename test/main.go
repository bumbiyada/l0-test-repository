package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	ttlcache "github.com/ReneKroon/ttlcache/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

var schema = `
CREATE TABLE aboba (
    idx text,
    msg text,
);`

type MessageDB struct {
	Idx string `db:"idx"`
	Msg string `db:"msg"`
}

func main() {
	// db

	db, err := sqlx.Connect("postgres", "user=foo dbname=bar sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv
	db.MustExec(schema)
	tx := db.MustBegin()
	tx.MustExec("INSERT INTO aboba (idx, msg) VALUES ($1, $2)", "idx1", "Moiron jmoiron@jmoiron.net")
	//nats
	opts := []nats.Option{nats.Name("NATS Streaming Example Subscriber")}
	nc, err := nats.Connect(stan.DefaultNatsURL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	sc, err := stan.Connect("test-cluster", "clientId")

	if err != nil {
		fmt.Println("err")
	}
	sub, err := sc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))

	}, stan.DeliverAllAvailable())

	time.Sleep(2 * time.Second)
	//
	//sc.Close()

	var cache ttlcache.SimpleCache = ttlcache.NewCache()

	var notFound = ttlcache.ErrNotFound
	cache.SetTTL(time.Duration(0))

	cache.Set("MyKey", "MyValue")
	cache.Set("MyNumber", 1000)

	val, err := cache.Get("MyKey2")
	if err == notFound {
		log.Fatalf("%s %T  %T", err, err, val)
	}
	cache.Remove("MyNumber")

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Printf("\nReceived an interrupt, unsubscribing and closing connection...\n\n")
			// Do not unsubscribe a durable on exit, except if asked to.

			sub.Unsubscribe()

			sc.Close()
			cache.Close()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
