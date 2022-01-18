package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	basics "main-app/pkg/basics"
	cfg "main-app/pkg/config"

	ttlcache "github.com/ReneKroon/ttlcache/v2"
	sqlx "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

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
type Mystruct struct {
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

// structure to transport  message to Cache in chanel
type keyval struct {
	Key string
	Val []byte
}

// structure to answer http listener about Cache data
type Response struct {
	Resp interface{} `json:"body"`
	Err  bool        `json:"error"`
}

// db constants
var schema = `
CREATE TABLE IF NOT EXISTS message  (
	order_uid varchar(20) PRIMARY KEY,
	track_number varchar(20),
	entry varchar(20),
	locale varchar(5),
	internal_signature varchar(20),
	customer_id varchar(20),
	delivery_service varchar(20),
	shardkey varchar(20),
	sm_id int4,
	date_created varchar(32),
	oof_shard varchar(20)
  );
  
  CREATE TABLE IF NOT EXISTS item (
	chrt_id int8 PRIMARY KEY,
	track_number varchar(20),
	price int4,
	rid varchar(20),
	name varchar(20),
	sale int4,
	size varchar(20),
	total_price int4,
	nm_id int8,
	brand varchar(20),
	status int2,
	fk_items varchar(20) REFERENCES message(order_uid)
  );
  CREATE TABLE IF NOT EXISTS payment (
	payment_id SERIAL PRIMARY KEY,
	transaction_fk varchar(20) REFERENCES message(order_uid),
	request_id varchar(20),
	currency varchar(20),
	provider varchar(20),
	amount int4,
	payment_dt int8,
	bank varchar(20),
	delivery_cost int4,
	goods_total int4,
	custom_fee int4
  );
  
  
  CREATE TABLE IF NOT EXISTS delivery (
	delivery_id SERIAL PRIMARY KEY,
	name varchar(64),
	phone varchar(16),
	zip int4,
	city varchar(32),
	address varchar(64),
	region varchar(32),
	email varchar(64),
	fk_delivery varchar(20) REFERENCES message(order_uid)
  );
`

const url = "postgres://postgres:123@db:5432/test?sslmode=disable"
const nats_url = "nats://nats-server:4222"

// Global channels
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

// init function
func init() {

	log.Println("...Starting initialization...")
	cache.SetTTL(time.Duration(0))
	log.Println(cfg.ABOBA)
}

// listener of signal interrupt
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

// NATS channel listener
func NATSListener() {
	var ClienID = "stan-sub"

	// simple con to nats,stan then get msg
	opts := []nats.Option{nats.Name("NATS Streaming Example Subscriber")}
	nc, err := nats.Connect("nats://nats-server:4222", opts...)
	basics.CheckErr(err, "failed to connect to nats")
	defer nc.Close()
	// stan.NatsURL(nats_url)
	log.Printf("[NATS-LISTENER] ClienID = %s URL = %s\n", ClienID, stan.NatsConn(nc))
	sc, err := stan.Connect("test-cluster", ClienID, stan.NatsConn(nc))
	basics.CheckErr(err, "failed to connect to stan")
	sub, err := sc.Subscribe("foo", func(m *stan.Msg) {
		log.Printf("[NATS] Received a msg: %23s\n", string(m.Data))
		NatsChan <- m.Data
	}, stan.DeliverAllAvailable())

	if true == <-DropAllChan {
		sub.Unsubscribe()
		sc.Close()
	}
}

// HTTP Listener function
func HttpListener() {

	type Req struct {
		Id string `json:"id"`
	}

	id := Req{}
	log.Printf("[HTTP] HOST %s PORT %s", cfg.HTTP_HOST, cfg.HTTP_PORT)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			r.ParseForm()
			log.Printf("[HTTP] GET REQUEST FROM CLIENT %s", r.Form)

			for k, _ := range r.Form {
				key := k
				e := json.Unmarshal([]byte(key), &id)
				basics.CheckErr(e, "Err while decoding id request from server func httplistener")
			}
			var ids string = id.Id
			GetCacheChan <- ids
			resp := <-ResponseHttpChan
			msg, err := json.Marshal(resp)
			basics.CheckErr(err, "Encoding msg for responce / func HTTP Listener")
			if resp.Err == true {
				log.Printf("[HTTP] %s VALUE is NOT in CACHE, sending this info to client", ids)
				log.Println("[HTTP] MESSAGE = ", string(msg))
				fmt.Fprintf(w, "%s", msg)
			} else {
				log.Printf("[HTTP] %s VALUE is in CACHE, sending this value to client", ids)
				log.Println("[HTTP] MESSAGE = ", string(msg))
				fmt.Fprintf(w, "%s", msg)
			}
		} else {
			log.Println("NOT POST REQUEST")
			fmt.Fprintf(w, "%s", "WHO ARE YOU ?")
		}
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.HTTP_HOST, cfg.HTTP_PORT), nil))

}

// Function that recovers Cache from DB, executed at start
func RecoverCache(Db *sqlx.DB) {
	const query string = `SELECT message.order_uid, message.track_number, entry, locale,
	internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard,
	request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee,
	delivery.name, phone, zip, city, address, region, email FROM message
	JOIN delivery ON message.order_uid = delivery.fk_delivery
	JOIN payment ON message.order_uid = payment.transaction_fk
	WHERE message.order_uid = $1;`

	const query_items string = `SELECT chrt_id, track_number, price, rid, name, sale, size,
	total_price, nm_id, brand, status FROM item WHERE fk_items = $1`
	type RecoverDb struct {
		Order_uid          string `db:"order_uid"`
		Track_number       string
		Entry              string
		Locale             string
		Internal_signature string
		Customer_id        string
		Delivery_service   string
		Shardkey           string
		Sm_id              int32
		Date_created       string
		Oof_shard          string
		Request_id         string
		Currency           string
		Provider           string
		Amount             int32
		Payment_dt         int64
		Bank               string
		Delivery_cost      int32
		Goods_total        int32
		Custom_fee         int32
		Name               string
		Phone              string
		Zip                int32
		City               string
		Address            string
		Region             string
		Email              string
	}
	var identifiers []string
	var tmp RecoverDb
	var tmp2 Mystruct
	var arr []items
	var cached keyval
	var err error
	// GET ALL ID`s FROM DB to iterate by them
	err = Db.Select(&identifiers, "SELECT order_uid FROM message;")
	basics.CheckErr(err, "error while Select Identifiers from db.Message   func = RecoverDb")
	// Iterating
	for idx, uid := range identifiers {
		log.Printf("\t\t[RECOVERING CACHE] \t [%v]/[%v]: [ID] = %v", (idx + 1), len(identifiers), uid)
		// GET data from db.Message, db.Delivery, db.Payment for single ID`s
		err = Db.Get(&tmp, query, uid)
		basics.CheckErr(err, "error while Select message/delivery/payment   func = RecoverDb")
		arr = nil
		err = Db.Select(&arr, query_items, uid)
		basics.CheckErr(err, "error while Select items from db.Item   func = RecoverDb")
		// SET DATA
		tmp2.Order_uid = tmp.Order_uid
		tmp2.Track_number = tmp.Track_number
		tmp2.Entry = tmp.Entry
		tmp2.Locale = tmp.Locale
		tmp2.Internal_signature = tmp.Internal_signature
		tmp2.Customer_id = tmp.Customer_id
		tmp2.Delivery_service = tmp.Delivery_service
		tmp2.Shardkey = tmp.Shardkey
		tmp2.Sm_id = tmp.Sm_id
		tmp2.Date_created = tmp.Date_created
		tmp2.Oof_shard = tmp.Oof_shard
		tmp2.Payment.Transactrion = tmp.Order_uid
		tmp2.Payment.Request_id = tmp.Request_id
		tmp2.Payment.Currency = tmp.Currency
		tmp2.Payment.Provider = tmp.Provider
		tmp2.Payment.Amount = tmp.Amount
		tmp2.Payment.Payment_dt = tmp.Payment_dt
		tmp2.Payment.Bank = tmp.Bank
		tmp2.Payment.Delivery_cost = tmp.Delivery_cost
		tmp2.Payment.Goods_total = tmp.Goods_total
		tmp2.Payment.Custom_fee = tmp.Custom_fee
		tmp2.Delivery.Name = tmp.Name
		tmp2.Delivery.Phone = tmp.Phone
		tmp2.Delivery.Zip = tmp.Zip
		tmp2.Delivery.City = tmp.City
		tmp2.Delivery.Address = tmp.Address
		tmp2.Delivery.Region = tmp.Region
		tmp2.Delivery.Email = tmp.Email
		tmp2.Items = arr
		// Convert to JSON and add data to Cache
		cached.Key = tmp2.Order_uid
		cached.Val, err = json.Marshal(tmp2)
		cache.Set(cached.Key, string(cached.Val))
	}
}

// Function to add value to DB
func AddToDB(Db *sqlx.DB, msg Mystruct) {
	tx := Db.MustBegin()
	tx.MustExec("INSERT INTO message (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", msg.Order_uid, msg.Track_number, msg.Entry, msg.Locale, msg.Internal_signature, msg.Customer_id, msg.Delivery_service, msg.Shardkey, msg.Sm_id, msg.Date_created, msg.Oof_shard)
	tx.MustExec("INSERT INTO payment (transaction_fk, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", msg.Payment.Transactrion, msg.Payment.Request_id, msg.Payment.Currency, msg.Payment.Provider, msg.Payment.Amount, msg.Payment.Payment_dt, msg.Payment.Bank, msg.Payment.Delivery_cost, msg.Payment.Goods_total, msg.Payment.Custom_fee)
	tx.MustExec("INSERT INTO delivery (name, phone, zip, city, address, region, email, fk_delivery) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", msg.Delivery.Name, msg.Delivery.Phone, msg.Delivery.Zip, msg.Delivery.City, msg.Delivery.Address, msg.Delivery.Region, msg.Delivery.Email, msg.Order_uid)
	ItemArr := msg.Items
	for _, item := range ItemArr {
		tx.MustExec("INSERT INTO item (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, fk_items) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)", item.Chrt_id, item.Track_number, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.Total_price, item.Nm_id, item.Brand, item.Status, msg.Order_uid)
	}
	tx.Commit()
	//Db.MustExec("INSERT INTO main (idx, texty) VALUES ($1, $2)", msg.Id, msg.Val)
	log.Println("\t\t [DB]\t\tadded to DB :", msg.Order_uid)
}

// Function that listens message from Nats func and decide what to do with message
func MessageHandler(Db *sqlx.DB) {

	for {
		select {
		case msg := <-NatsChan:
			var res Mystruct
			err := json.Unmarshal(msg, &res)
			basics.CheckErr(err, "failed to decode msg / func MessageHandler")
			// log.Println(res.Order_uid, string(msg))
			ValidCacheChan <- res.Order_uid
			resp := <-ValidRespCacheChan
			if resp == true {
				log.Println("\t\t[MSG] msg allready in Cache :", res.Order_uid)
			} else {
				log.Printf("\t\t[MSG] msg NOT in cache, so let`s add it id = %s", res.Order_uid)
				var cache keyval
				cache.Key = res.Order_uid
				cache.Val = msg
				AddCacheChan <- cache
				log.Printf("\t\t[MSG] also let`s add to DB, id = %s", res.Order_uid)
				AddToDB(Db, res)
			}
		}
	}
}

// Function that manage Cache
func CacheHandler() {

	for {
		select {
		case id := <-GetCacheChan:
			//get val by id to http
			val, e := cache.Get(id)
			err := false
			if e == notFound {
				err = true
			}
			log.Println("[CACHE] GOT FROM HTTP-server ", id)
			var resp Response
			resp.Resp = val
			resp.Err = err
			ResponseHttpChan <- resp
		case cacheItem := <-AddCacheChan:
			// add to cache
			cache.Set(cacheItem.Key, string(cacheItem.Val))
			log.Println("\t\t[CACHE]\t ADDED VALUE ", cacheItem.Key)
		case id := <-ValidCacheChan:
			var resp bool
			_, e := cache.Get(id)
			if e == notFound {
				resp = false
				log.Println("\t\t[CACHE]\tDidn`t Find in cache val by id = ", id)
			} else {
				resp = true
				log.Println("\t\t[CACHE]\tFound in cache val by id = ", id)
			}
			ValidRespCacheChan <- resp
		case <-DropAllChan:
			cache.Purge()
			cache.Close()
			return
		default:
			//skip
		}
	}
}

// MAIN function
func main() {

	Db, err := sqlx.Connect("postgres", url)
	basics.CheckErr(err, "error while connecting to Database")
	// init db
	Db.MustExec(schema)

	// RESTORE CACHE FROM DB
	var WG_restoreCache sync.WaitGroup
	WG_restoreCache.Add(1)
	go func() {
		log.Println("\t\tRestoring Cache from DB ...")
		RecoverCache(Db)
		WG_restoreCache.Done()
	}()
	WG_restoreCache.Wait()
	//wg for every single goroutine
	var WG_Nats sync.WaitGroup
	var WG_Http sync.WaitGroup
	var WG_Msg sync.WaitGroup
	var WG_Cache sync.WaitGroup
	var WG_SigInt sync.WaitGroup

	// STARTING GOROUTINES
	WG_Nats.Add(1)
	go func() {
		log.Println("\t\tStarting NATS listener ...")
		NATSListener()
		WG_Nats.Done()
	}()
	WG_Http.Add(1)
	go func() {
		log.Println("\t\tStarting Http listener ...")
		HttpListener()
		WG_Http.Done()
	}()
	WG_Msg.Add(1)
	go func() {
		log.Println("\t\tStarting Message handler ...")
		MessageHandler(Db)
		WG_Msg.Done()
	}()
	WG_Cache.Add(1)
	go func() {
		log.Println("\t\tStarting Cache handler ...")
		CacheHandler()
		WG_Cache.Done()
	}()
	WG_SigInt.Add(1)
	go func() {
		log.Println("\t\tStarting interrupt LISTENER ...")
		SigIntListener()
		WG_SigInt.Done()
	}()
	log.Println("\t\tAll Goroutines started..")
	//time.Sleep(time.Second * 1)
	WG_Nats.Wait()
	WG_Http.Wait()
	WG_Msg.Wait()
	WG_Cache.Wait()

}
