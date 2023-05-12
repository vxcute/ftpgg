package main

import (
	"fmt"
	"goftp/ftpgg"
	"log"
)

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
		fmt.Println(d)
	}

	pwd, err := ftp.Pwd() 

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(pwd)
}