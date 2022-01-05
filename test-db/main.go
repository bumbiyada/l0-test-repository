package main

import (
	"fmt"
	"log"

	sqlx "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const conf = "host=localhost port=5432 user=postgres password=123 dbname=test sslmode=disable"

var schema = `
CREATE TABLE main (
	idx text,
	text text
);
`

type Mystruct struct {
	Idx  string
	Text string
}

func main() {
	fmt.Println("starting my application")
	db, err := sqlx.Connect("postgres", conf)
	CheckErr(err, "error while connecting to database")
	// init db
	db.MustExec(schema)
	// add values
	db.MustExec(fmt.Sprintf("INSERT INTO main (idx, text) VALUES (%s, %s)", "1", "text1"))
	db.MustExec(fmt.Sprintf("INSERT INTO main (idx, text) VALUES (%s, %s)", "2", "text2"))
	db.MustExec(fmt.Sprintf("INSERT INTO main (idx, text) VALUES (%s, %s)", "3", "text3"))
	db.MustExec(fmt.Sprintf("INSERT INTO main (idx, text) VALUES (%s, %s)", "4", "aboba1"))
	db.MustExec(fmt.Sprintf("INSERT INTO main (idx, text) VALUES (%s, %s)", "5", "aboba2"))
	db.MustExec(fmt.Sprintf("INSERT INTO main (idx, text) VALUES (%s, %s)", "6", "aboba3"))
	// get values
	array := []Mystruct{}

	db.Select(&array, "SELECT * FROM main ORDER BY idx ASC")

	for _, val := range array {
		log.Println(val)
	}
}

func CheckErr(e error, description string) {
	if e != nil {
		log.Fatalf("[ERROR]: %s\n[INFO]: %s", e.Error(), description)
	}
}
