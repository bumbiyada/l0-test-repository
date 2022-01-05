package main

import (
	"fmt"
	"os"
	"os/signal"

	// "encoding/json"
	"context"
	"flag"
	"log"

	"sync"
	"time"

	// stan "github.com/nats-io/stan.go"
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

type Sub struct {
	clusterID, clientID string
	URL                 string
	userCreds           string
	showTime            bool
	qgroup              string
	unsubscribe         bool
	startSeq            uint64
	startDelta          string
	deliverAll          bool
	newOnly             bool
	deliverLast         bool
	durable             string
}

var usageStr = `
Usage of this app

Options:

see you next time
`
var wg sync.WaitGroup

// Function that listen interrupt signal and close App
func handle_cancel(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal)
	dropSub := make(chan bool)
	signal.Notify(sigCh, os.Interrupt)
	for {
		sig := <-sigCh
		switch sig {
		case os.Interrupt:
			cancel()
			dropSub <- true
			return
		}
	}
}

// Main function of Nats subscriber
func sub_listener(sub_flags *Sub, dropSub chan bool) {
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
	})

	fmt.Println(sub_flags)
	time.Sleep(2 * time.Second)
	//
	go func() {
		for range dropSub {
			a := <-dropSub
			if a == true {
				fmt.Printf("\nReceived an interrupt, unsubscribing and closing connection...\n\n")
				sub.Unsubscribe()

				sc.Close()
			}

			// Do not unsubscribe a durable on exit, except if asked to.

		}
	}()
}

// Main function of Http server
func http_listener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Subscriber listener Closed in context")
			return
		default:
			time.Sleep(2 * time.Second)
			fmt.Println("http listener works fine...")
		}
	}
}

// Function that recover cash from DB
func recover_cash() {
	fmt.Println("cash recovered successfully")
}

// Function that describe app, starts if app executes without args
func usage() {
	log.Fatalf(usageStr)
}

// Function to describe state of app
func state() {
	fmt.Println("describe current state")
	fmt.Println(time.Now())

}

// Main function
func main() {
	var (
		cash_size    int
		recover_cash bool
	)
	// my flags
	flag.BoolVar(&recover_cash, "recover_cash", false, "recover cash from DB at start")
	flag.IntVar(&cash_size, "cash", 1000, "size of cash")

	var sub_flags Sub
	flag.StringVar(&sub_flags.URL, "s", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&sub_flags.URL, "server", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&sub_flags.clusterID, "c", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&sub_flags.clusterID, "cluster", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&sub_flags.clientID, "id", "stan-sub", "The NATS Streaming client ID to connect with")
	flag.StringVar(&sub_flags.clientID, "clientid", "stan-sub", "The NATS Streaming client ID to connect with")
	flag.BoolVar(&sub_flags.showTime, "t", false, "Display timestamps")
	// Subscription options
	flag.Uint64Var(&sub_flags.startSeq, "seq", 0, "Start at sequence no.")
	flag.BoolVar(&sub_flags.deliverAll, "all", true, "Deliver all")
	flag.BoolVar(&sub_flags.newOnly, "new_only", false, "Only new messages")
	flag.BoolVar(&sub_flags.deliverLast, "last", false, "Start with last value")
	flag.StringVar(&sub_flags.startDelta, "since", "", "Deliver messages since specified time offset")
	flag.StringVar(&sub_flags.durable, "durable", "", "Durable subscriber name")
	flag.StringVar(&sub_flags.qgroup, "qgroup", "", "Queue group name")
	flag.BoolVar(&sub_flags.unsubscribe, "unsub", false, "Unsubscribe the durable on exit")
	flag.BoolVar(&sub_flags.unsubscribe, "unsubscribe", false, "Unsubscribe the durable on exit")
	flag.StringVar(&sub_flags.userCreds, "cr", "", "Credentials File")
	flag.StringVar(&sub_flags.userCreds, "creds", "", "Credentials File")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		log.Printf("Error: A subject must be specified.")
		usage()
	}

	state()

	ctx, cancel := context.WithCancel(context.Background())

	go handle_cancel(cancel)
	wg.Add(1)
	go func() {
		sub_listener(&sub_flags, dropSub)

		wg.Done()
	}()
	wg.Add(1)
	go func() {
		http_listener(ctx)

		wg.Done()
	}()
	fmt.Println("goroutine started")
	wg.Wait()
	fmt.Println("app finished")
}
