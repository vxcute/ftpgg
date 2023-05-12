package main

import (
	"bytes"
	"fmt"
	"goftp/ftpgg"
	"io"
	"log"
	"os"
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

	file, err := ftp.Download("file.txt")

	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("file.txt") 

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	if _, err := io.Copy(f, bytes.NewReader(file)); err != nil {
		log.Fatal(err)
	}
}