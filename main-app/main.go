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
	Delivery_servise   string   `json:"delivery_servise"`
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
	resp interface{}
	err  bool
}

// db constants
var schema = `
CREATE TABLE IF NOT EXISTS main (
	idx text,
	texty text
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

}

// function that log error  if error != nil
func CheckErr(e error, description string) {

	if e != nil {
		log.Fatalf("\t[ERROR]: \t%s\n\t\t\t[DESCRIPTION]: \t%s", e.Error(), description)
	}
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

	// simple con to nats,stan then get msg
	opts := []nats.Option{nats.Name("NATS Streaming Example Subscriber")}
	nc, err := nats.Connect("nats://nats-server:4222", opts...)
	CheckErr(err, "failed to connect to nats")
	defer nc.Close()
	sc, err := stan.Connect("test-cluster", "clientid", stan.NatsURL(nats_url))
	CheckErr(err, "failed to connect to stan")
	sub, err := sc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
		NatsChan <- m.Data
	}, stan.DeliverAllAvailable())

	if true == <-DropAllChan {
		sub.Unsubscribe()
		sc.Close()
	}

}

// HTTP Listener function
func HttpListener() {

	http.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		value := r.FormValue("MyValue")
		fmt.Fprintf(w, "Got value from front %s\n", value)
		GetCacheChan <- value
		resp := <-ResponseHttpChan
		fmt.Println("rep", resp)
		if resp.err == true {
			fmt.Fprintf(w, "No data in cache by id : %s\n", value)
		} else {
			fmt.Fprintf(w, "Here is your data sir...   %s\n", resp.resp)
		}
		// 	data := SomeStruct{}
		// w.Header().Set("Content-Type", "application/json")
		// w.WriteHeader(http.StatusCreated)
		// json.NewEncoder(w).Encode(data)
	})
	log.Fatal(http.ListenAndServe(":3000", nil))

	// log.Println("...STARTING SERVER...")

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	//json.NewEncoder(w).Encode(data)
	// 	//log.Println(r)

	// 	// fmt.Fprintf(w, "Successfully Uploaded Value %s\n", value)
	// 	r.ParseForm()
	// 	// log.Println("POST FORM")
	// 	// log.Println(r.PostForm)
	// 	log.Println("FORM")
	// 	log.Println(r.Form)
	// 	//json.Unmarshal([]byte(r.Form[0][0]), &data2)
	// 	if r.Method == http.MethodPost {
	// 		log.Println("FUCK YOU")
	// 	}
	// 	//r.ParseForm()       //анализ аргументов,
	// 	//fmt.Println(r.Form) // ввод информации о форме на стороне сервера
	// 	//fmt.Println("path", r.URL.Path)
	// 	//fmt.Println("scheme", r.URL.Scheme)
	// 	//fmt.Println(r.Form["url_long"])
	// 	for k, _ := range r.Form {
	// 		fmt.Println("key:", k)
	// 		key := k
	// 		log.Println(r.PostFormValue(key))
	// 		e := json.Unmarshal([]byte(key), &data2)
	// 		if e != nil {
	// 			log.Fatal("aboba")
	// 		}
	// 	}
	// 	log.Printf("[ID] = %v", data2.Id)
	// 	data := Mystruct{data2.Id, data2.Id}
	// 	JS, err := json.Marshal(data)
	// 	if err != nil {
	// 		log.Fatalf("errror %s", err)
	// 	}
	// 	fmt.Fprintf(w, string(JS)) // отправляем данные на клиентскую сторону

	// 	//log.Println(r)
	// 	//w.Header().Set("Content-Type", "application/json")

	// 	// w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, DELETE, PUT")
	// 	// w.Header().Set("Access-Control-Allow-Headers", "content-type")
	// 	// w.Header().Set("Access-Control-Max-Age", "86400")
	// 	// w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	// 	w.WriteHeader(http.StatusCreated)
	// 	// log.Println(http.StatusCreated)
	// 	// log.Println(w)

	// 	// func GetPeopleAPI(w http.ResponseWriter, r *http.Request) {

	// 	// 	//Allow CORS here By * or specific origin
	// 	// 	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 	// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	// 	// 	// return "OKOK"
	// 	// 	json.NewEncoder(w).Encode("OKOK")
	// 	// }
	// 	// router := mux.NewRouter()
	// 	// //json.NewEncoder(w).Encode(data)
	// 	// router.HandleFunc("/api", GetPeopleAPI).Methods("POST", "OPTIONS")
	// })

	// log.Fatal(http.ListenAndServe("localhost:8080", nil))

}

// Function that recovers Cache from DB, executed at start
func RecoverCache(Db *sqlx.DB) {

	array := []Mystruct{}
	var tmp Mystruct
	Db.Select(&array, "SELECT * FROM main ORDER BY idx ASC")
	for idx, val := range array {
		json.Marshal(val)
		log.Printf("recovered %v : %v", idx, val)
		cache.Set(val.Id, val.Val)
	}
}

// Function to add value to DB
func AddToDB(Db *sqlx.DB, msg Mystruct) {
	tx := Db.MustBegin()
	tx.MustExec("INSERT INTO message (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", msg.Order_uid, msg.Track_number, msg.Entry, msg.Locale, msg.Internal_signature, msg.Customer_id, msg.Delivery_servise, msg.Shardkey, msg.Sm_id, msg.Date_created, msg.Oof_shard)
	tx.MustExec("INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_id, bank, delivery_cost, goods_total, custom_fee, fk_payment) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", msg.Payment.Transactrion, msg.Payment.Request_id, msg.Payment.Currency, msg.Payment.Provider, msg.Payment.Amount, msg.Payment.Bank, msg.Payment.Delivery_cost, msg.Payment.Goods_total, msg.Payment.Custom_fee, msg.Order_uid)
	tx.MustExec("INSERT INTO delivery (name, phone, zip, city, address, region, email, fk_delivery) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", msg.Delivery.Name, msg.Delivery.Phone, msg.Delivery.Zip, msg.Delivery.City, msg.Delivery.Address, msg.Delivery.Region, msg.Delivery.Email, msg.Order_uid)
	ItemArr := msg.Items
	for _, item := range ItemArr {
		tx.MustExec("INSERT INTO item (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, fk_item) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)", item.Chrt_id, item.Track_number, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.Total_price, item.Nm_id, item.Brand, item.Status, msg.Order_uid)
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
			CheckErr(err, "failed to decode msg / func MessageHandler")
			log.Println(res.Order_uid, string(msg))
			ValidCacheChan <- res.Order_uid
			resp := <-ValidRespCacheChan
			if resp == false {
				fmt.Println("msg allready in Cache / func MessageHandler")
			} else {
				var cache keyval
				cache.Key = res.Order_uid
				cache.Val = msg
				AddCacheChan <- cache
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
			var resp Response
			resp.resp = val
			resp.err = err
			ResponseHttpChan <- resp
		case cacheItem := <-AddCacheChan:
			// add to cache
			cache.Set(cacheItem.Key, cacheItem.Val)
			log.Println("added ", cacheItem.Key, cacheItem.Val)
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
	CheckErr(err, "error while connecting to Database")
	// init db
	Db.MustExec(schema)

	// RESTORE CACHE FROM DB
	var WG_restoreCache sync.WaitGroup
	WG_restoreCache.Add(1)
	go func() {
		log.Println("Restoring Cache from DB ...")
		RecoverCache(Db)
		WG_restoreCache.Done()
	}()

	//wg for every single goroutine
	var WG_Nats sync.WaitGroup
	var WG_Http sync.WaitGroup
	var WG_Msg sync.WaitGroup
	var WG_Cache sync.WaitGroup
	var WG_SigInt sync.WaitGroup

	// STARTING GOROUTINES
	WG_Nats.Add(1)
	go func() {
		log.Println("Starting NATS listener ...")
		NATSListener()
		WG_Nats.Done()
	}()
	WG_Http.Add(1)
	go func() {
		log.Println("Starting Http listener ...")
		NATSListener()
		WG_Http.Done()
	}()
	WG_Msg.Add(1)
	go func() {
		log.Println("Starting Message handler ...")
		MessageHandler(Db)
		WG_Msg.Done()
	}()
	WG_Cache.Add(1)
	go func() {
		log.Println("Starting Cache handler ...")
		CacheHandler()
		WG_Cache.Done()
	}()
	WG_SigInt.Add(1)
	go func() {
		log.Println("Starting interrupt LISTENER ...")
		SigIntListener()
		WG_SigInt.Done()
	}()
	log.Println("All Goroutines started..")
	time.Sleep(time.Second * 1)
}
