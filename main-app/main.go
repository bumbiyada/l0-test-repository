package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	ttlcache "github.com/ReneKroon/ttlcache/v2"
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

type Mystruct struct {
	Id  string
	Val string
}
type keyval struct {
	Key string
	Val []byte
}

type Response struct {
	resp string
	err  bool
}

var GetCacheChan = make(chan string)
var AddCacheChan = make(chan keyval)
var ResponseHttpChan = make(chan Response)
var NatsChan = make(chan []byte)
var ValidCacheChan = make(chan string)
var ValidRespCacheChan = make(chan bool)
var DropAllChan = make(chan bool)

// cache
var cache ttlcache.SimpleCache = ttlcache.NewCache()
var notFound = ttlcache.ErrNotFound

func init() {
	fmt.Println("...Starting initialization...")
	cache.SetTTL(time.Duration(0))
}

func SigIntListener() {
	//if sigint then close all will be added
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	for {
		sig := <-sigCh
		switch sig {
		case os.Interrupt:
			DropAllChan <- true
			return
		}
	}
}

func NATSListener() {
	// simple con to nats,stan then get msg
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
	//chan <- msg
	var message Mystruct
	message.Id = "1"
	message.Val = "aboba"
	msg, err := json.Marshal(message)
	if err != nil {
		log.Fatalln("error encoding")
	}
	NatsChan <- msg
	time.Sleep(time.Second * 5)
	if true == <-DropAllChan {
		sub.Unsubscribe()
		sc.Close()
	}

}

func HttpListener() {
	// listen / GET <id>
	// chan <- id
	var id string
	GetCacheChan <- id
	resp := <-ResponseHttpChan
	fmt.Println("rep", resp)

}

func RecoverCache() {
	fmt.Println("recovered Cache")
}

func AddToDB(msg Mystruct) {
	fmt.Println("added to DB ", msg)
}
func MessageHandler() {
	for {
		select {
		case msg := <-NatsChan:
			var res Mystruct
			err := json.Unmarshal(msg, &res)
			if err != nil {
				log.Println("Failed while decoding message..")
			}
			log.Println(res.Id, res.Val)
			ValidCacheChan <- res.Id
			resp := <-ValidRespCacheChan
			if resp == false {
				fmt.Println("msg allready in Cache")
			} else {
				var cache keyval
				cache.Key = res.Id
				cache.Val = msg
				AddCacheChan <- cache
				AddToDB(res)
			}
		}
	}
}

func CacheHandler() {
	for {
		select {
		case id := <-GetCacheChan:
			//get val by id to http
			val, eror := cache.Get(id)
			err := false
			if eror == notFound {
				err = true
			}
			var resp Response
			resp.resp = val
			resp.err = err
			ResponseHttpChan <- resp
		case cacheVal := <-AddCacheChan:
			// add to cache
			cache.Set(cacheVal.Key, cacheVal.Val)
			log.Println("added ", cache.Key, cache.Val)
		case id := <-ValidCacheChan:
			var resp bool
			_, eror := cache.Get(id)
			if eror == notFound {
				resp = false
				log.Println("Didn`t Find in cache val by id = ", id)
			} else {
				resp = true
				log.Println("Found in cache val by id = ", id)
			}
			ValidRespCacheChan <- resp
		case <-CloseChan:
			cache.Purge()
			cache.Close()
			return
		default:
			//skip
		}
	}
}

func main() {
	RecoverCache() //done
	// config
	//wg for every single goroutine
	var WG_Nats sync.WaitGroup
	var WG_Http sync.WaitGroup
	var WG_Msg sync.WaitGroup
	var WG_Cache sync.WaitGroup
	var WG_SigInt sync.WaitGroup
	WG_Nats.Add(1)
	go func() {
		NATSListener()
		WG_Nats.Done()
	}()
	WG_Http.Add(1)
	go func() {
		NATSListener()
		WG_Http.Done()
	}()
	WG_Msg.Add(1)
	go func() {
		MessageHandler()
		WG_Msg.Done()
	}()
	WG_Cache.Add(1)
	go func() {
		CacheHandler()
		WG_Cache.Done()
	}()
	WG_SigInt.Add(1)
	go func() {
		SigIntListener()
		WG_SigInt.Done()
	}()
	log.Println("All Goroutines started..")
	time.Sleep(time.Second * 1)
}
