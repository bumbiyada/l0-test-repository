// Copyright 2016-2019 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

var usageStr = `
Usage: stan-pub [options] <subject> <message>

Options:
	-s,  --server   <url>            NATS Streaming server URL(s)
	-c,  --cluster  <cluster name>   NATS Streaming cluster name
	-id, --clientid <client ID>      NATS Streaming client ID
	-a,  --async                     Asynchronous publish mode
	-g	 --generator				 Generated values
	-cr, --creds    <credentials>    NATS 2.0 Credentials
`

// NOTE: Use tls scheme for TLS, e.g. stan-pub -s tls://demo.nats.io:4443 foo hello
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

// Function that generates random string of fixed len
func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Function that generate message in struct, then parse it to json and msg -> main
func generateMsg() []byte {
	// Struct of delivery Json
	type delivery struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Zip     int32  `json:"zip"`
		City    string `json:"city"`
		Address string `json:"address"`
		Region  string `json:"region"`
		Email   string `json:"email"`
	}
	// Struct of PAyment Json
	type payment struct {
		Transactrion  string `json:"transactrion"`
		Request_id    string `json:"request_id"`
		Currency      string `json:"currency"`
		Provider      string `json:"provider"`
		Amount        int32  `json:"amount"`
		Payment_dt    int64  `json:"payment_dt"`
		Bank          string `json:"bank"`
		Delivery_cost int32  `json:"delivery_cost"`
		Goods_total   int32  `json:"goods_total"`
		Custom_fee    int32  `json:"custom_fee"`
	}
	// Struct of Items Json
	type items struct {
		Chrt_id      int64  `json:"chrt_id"`
		Track_number string `json:"track_number"`
		Price        int32  `json:"price"`
		Rid          string `json:"rid"`
		Name         string `json:"name"`
		Sale         int32  `json:"sale"`
		Size         string `json:"size"`
		Total_price  int32  `json:"total_price"`
		Nm_id        int64  `json:"nm_id"`
		Brand        string `json:"brand"`
		Status       int16  `json:"status"`
	}
	// Struct of All message together Json
	type message struct {
		Order_uid          string   `json:"order_uid"`
		Track_number       string   `json:"track_number"`
		Entry              string   `json:"entry"`
		Delivery           delivery `json:"delivery"`
		Payment            payment  `json:"payment"`
		Items              []items  `json:"items"`
		Locale             string   `json:"locale"`
		Internal_signature string   `json:"internal_signature"`
		Customer_id        string   `json:"customer_id"`
		Delivery_service   string   `json:"delivery_service"`
		Shardkey           string   `json:"shardkey"`
		Sm_id              int32    `json:"sm_id`
		Date_created       string   `json:"date_created"`
		Oof_shard          string   `json:"oof_shard"`
	}
	// define message except items
	PersonName := [...]string{"Ivan", "Petr", "John", "Valery", "Ryan", "Keanu", "Steve"}
	PersonSurname := [...]string{"Ivanov", "Petrov", "Sidorov", "Parker", "Gosling", "Reeves", "Buscemi"}
	CityName := [...]string{"Moscow", "NY", "Paris", "Voronezh", "Kyiv", "Tarkov"}
	AdressName := [...]string{"50letVLKSM", "Polyarnya", "Lenina", "Stalina", "avenue"}
	BankName := [...]string{"SberBank", "Alpha", "Betta", "Gamma", "JWB"}
	var NewMessage message
	NewMessage.Order_uid = randStringBytes(4) + "uid"
	NewMessage.Track_number = randStringBytes(10) + "track"
	NewMessage.Entry = "WBIL"
	NewMessage.Delivery.Name = PersonName[rand.Intn(len(PersonName))] + PersonSurname[rand.Intn(len(PersonSurname))]
	NewMessage.Delivery.Phone = "+7985" + strconv.Itoa(rand.Intn(9999999))
	NewMessage.Delivery.Zip = rand.Int31n(999999)
	NewMessage.Delivery.City = CityName[rand.Intn(len(CityName))]
	NewMessage.Delivery.Address = AdressName[rand.Intn(len(AdressName))] + " " + strconv.Itoa(rand.Intn(100))
	NewMessage.Delivery.Region = "Solar system/Earth"
	NewMessage.Delivery.Email = randStringBytes(8) + "@gmail.com"
	NewMessage.Payment.Transactrion = NewMessage.Order_uid
	NewMessage.Payment.Request_id = "req_id"
	NewMessage.Payment.Currency = "USD"
	NewMessage.Payment.Provider = "wbpay"
	NewMessage.Payment.Amount = rand.Int31n(3000)
	NewMessage.Payment.Payment_dt = rand.Int63n(9999999999)
	NewMessage.Payment.Bank = BankName[rand.Intn(len(BankName))]
	NewMessage.Payment.Delivery_cost = rand.Int31n(3000)
	NewMessage.Payment.Goods_total = NewMessage.Payment.Amount - NewMessage.Payment.Delivery_cost
	NewMessage.Payment.Custom_fee = 0
	NewMessage.Locale = "en"
	NewMessage.Internal_signature = "intern_sign"
	NewMessage.Customer_id = PersonName[rand.Intn(len(PersonName))] + PersonSurname[rand.Intn(len(PersonSurname))] + strconv.Itoa(rand.Intn(9999))
	NewMessage.Delivery_service = "delivery"
	NewMessage.Shardkey = "shardkey"
	NewMessage.Sm_id = 12
	NewMessage.Date_created = "DATA"
	NewMessage.Oof_shard = "abc"
	// decide how much items it will be (1 to 3) and define them
	var itemArr []items
	var item items
	//itemNbr := rand.Intn(2) + 1
	itemNbr := 1
	for i := 0; i < itemNbr; i++ {
		item.Chrt_id = rand.Int63n(9999999)
		item.Track_number = NewMessage.Track_number
		item.Price = rand.Int31n(30000)
		item.Rid = randStringBytes(15)
		item.Name = "ItemName"
		item.Sale = 0
		item.Size = "0"
		item.Total_price = item.Price
		item.Nm_id = rand.Int63n(9999999)
		item.Brand = "ItemBrand"
		item.Status = 202
		itemArr = append(itemArr, item)
	}
	// add items to message
	NewMessage.Items = itemArr
	// fmt.Println(NewMessage)
	// convert Message to Json
	JsonData, err := json.Marshal(NewMessage)
	if err != nil {
		log.Fatalf("error while encoding struct to Json : %s", err)
	}
	// fmt.Println(string(JsonData))
	return JsonData
}
func main() {
	// make some vars
	var (
		clusterID string
		clientID  string
		URL       string
		async     bool
		gen       bool
		userCreds string
	)
	// make some flags
	flag.StringVar(&URL, "s", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&URL, "server", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&clusterID, "c", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clusterID, "cluster", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clientID, "id", "stan-pub", "The NATS Streaming client ID to connect with")
	flag.StringVar(&clientID, "clientid", "stan-pub", "The NATS Streaming client ID to connect with")
	flag.BoolVar(&async, "a", false, "Publish asynchronously")
	flag.BoolVar(&async, "async", false, "Publish asynchronously")
	flag.BoolVar(&gen, "g", false, "Publish generated values")
	flag.BoolVar(&gen, "generator", false, "Publish generated values")
	flag.StringVar(&userCreds, "cr", "", "Credentials File")
	flag.StringVar(&userCreds, "creds", "", "Credentials File")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		usage()
	}

	subj, msg := args[0], []byte(args[1])
	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Streaming Example Publisher")}
	// Use UserCredentials
	if userCreds != "" {
		opts = append(opts, nats.UserCredentials(userCreds))
	}

	// Connect to NATS
	nc, err := nats.Connect(URL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	// Connect to STAN aka NATS streaming
	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}
	defer sc.Close()

	// make message and subject

	ch := make(chan bool)
	var glock sync.Mutex
	var guid string
	acb := func(lguid string, err error) {
		glock.Lock()
		log.Printf("Received ACK for guid %s\n", lguid)
		defer glock.Unlock()
		if err != nil {
			log.Fatalf("Error in server ack for guid %s: %v\n", lguid, err)
		}
		if lguid != guid {
			log.Fatalf("Expected a matching guid in ack callback, got %s vs %s\n", lguid, guid)
		}
		ch <- true
	}
	if !gen {
		if !async {
			err = sc.Publish(subj, msg)
			if err != nil {
				log.Fatalf("Error during publish: %v\n", err)
			}
			log.Printf("Published [%s] : '%s'\n", subj, msg)
		} else {
			glock.Lock()
			guid, err = sc.PublishAsync(subj, msg, acb)
			if err != nil {
				log.Fatalf("Error during async publish: %v\n", err)
			}
			glock.Unlock()
			if guid == "" {
				log.Fatal("Expected non-empty guid to be returned.")
			}
			log.Printf("Published [%s] : '%s' [guid: %s]\n", subj, msg, guid)

			select {
			case <-ch:
				break
			case <-time.After(5 * time.Second):
				log.Fatal("timeout")
			}

		}
	} else {
		var (
			i     int64
			count int64
		)
		count, parseError := strconv.ParseInt(string(msg), 10, 64)
		if parseError != nil {
			log.Printf("Error during parsing argv[1] to integer in generator publish: %v\n set deffault = 10", err)
			count = 10
		}
		for i = 0; i < count; i++ {
			msg := generateMsg()
			err = sc.Publish(subj, msg)
			if err != nil {
				log.Fatalf("Error during generator publish: %v\n", err)
			}
			log.Printf("Published generator %v [%s] : '%s'\n", i, subj, msg)
		}

	}
}
