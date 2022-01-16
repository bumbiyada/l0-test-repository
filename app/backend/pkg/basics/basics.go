package basics

import "log"

func CheckErr(e error, description string) {

	if e != nil {
		log.Fatalf("\t[ERROR]: \t%s\n\t\t\t[DESCRIPTION]: \t%s", e.Error(), description)
	}
}

func Aboba() {
	log.Println("ABOBA")
}
