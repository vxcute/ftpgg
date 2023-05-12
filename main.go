package main

import (
	"fmt"
	"goftp/ftpgg"
	"log"
)

func main() {	

	ftp := ftpgg.NewFTP(":")

	fmt.Println(ftp.Connect())

	fmt.Println(ftp.Login(ftpgg.FTPLogin{Username: "ftpuser", Password: "pass"}))

	dirs, err := ftp.List()

	if err != nil {
		log.Fatal(err)
	}

	for _, d := range dirs {
		fmt.Println(d)
	}
}