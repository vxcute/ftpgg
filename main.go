package main

import (
	"fmt"
	"goftp/ftpgg"
	"log"
	"time"
)

type EntryType int

const ( 
	RegularFile EntryType = iota
	Directory 
	Link
)

type Entry struct {
	Name 		string
	Date 		time.Time
	Permissions string
}

func ParseDate(d string) (time.Time, error) {
	layout := "Jan 2 15:04"
	return time.Parse(layout, d)
}

func main() {	

	ftp := ftpgg.NewFTP(":")

	r, err := ftp.Connect()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r)
	
	r, err = ftp.Login(ftpgg.FTPLogin{Username: "ftpuser", Password: "pass"})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r)

	dirs, err := ftp.List()

	if err != nil {
		log.Fatal(err)
	}

	for _, d := range dirs {
		fmt.Printf("%d - %s - %s - %s\n", d.Type, d.Name, d.Date.String(), d.Permissions)
	}
}