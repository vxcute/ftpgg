package main

import (
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
	Name        string
	Date        time.Time
	Permissions string
}

func ParseDate(d string) (time.Time, error) {
	layout := "Jan 2 15:04"
	return time.Parse(layout, d)
}

func main() {

	ftp := ftpgg.NewFTP(":")

	if err := ftp.Connect(); err != nil {
		log.Fatal(err)
	}

	if err := ftp.Login(ftpgg.FTPLogin{Username: "ftpuser", Password: "pass"}); err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := ftp.Quit(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := ftp.Cwd("mydir"); err != nil {
		log.Fatal(err)
	}

	if err := ftp.Cdup(); err != nil {
		log.Fatal(err)
	}

	if err := ftp.Chmod("file.txt", "x"); err != nil {
		log.Fatal(err)
	}
}
