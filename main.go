package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

const (
	ServiceReadyForNewUser   = "220"
	UserNameOkayNeedPassword = "331"
	LoginSuccessful 		 = "230"
	LoginIncorrect 			 = "530"
)

var ( 
	List = []byte("LIST\r\n")
	Pwd  = []byte("PWD\r\n")
	Espv = []byte("EPSV\r\n")
)

type FTP struct {
	addr string
	conn net.Conn
	controlBuf []byte
	dataBuf    []byte
}

type FTPLogin struct {
	username string
	password string
}

func NewFTP(addr string) *FTP {
	return &FTP{ 
		addr: 		addr,
		controlBuf: make([]byte, 255),
		dataBuf: 	make([]byte, 2048),
	}
}

func (f *FTP) Connect() (string, error) {

	var err error 

	f.conn, err = net.Dial("tcp", f.addr+"21")

	if err != nil {
		return "", err
	}

	n, err := f.conn.Read(f.controlBuf)

	if err != nil {
		return "", err
	}

	return string(f.controlBuf[:n-2]), nil
}

func (f *FTP) enterPassiveMode() (string, error) {

	_, err := f.conn.Write(Espv)

	if err != nil {
		return "", err
	}

	n, err := f.conn.Read(f.controlBuf)

	if err != nil {
		return "", err
	}

	resp := string(f.controlBuf[3:n])

	rgx := regexp.MustCompile("[0-9]+")

	return rgx.FindString(resp), nil
}

func (f *FTP) List() ([]string, error) {

	port, err := f.enterPassiveMode()

	if err != nil  {
		return nil, err
	}

	fmt.Println("port: ", port)

	dataConn, err := net.Dial("tcp", f.addr+port)

	if err != nil {
		return nil, err
	}

	defer dataConn.Close()

	_, err = f.conn.Write(List)

	if err != nil {
		return nil, err
	}

	dirs := make(chan string, 1)

	go func() {

		if err != nil {
			dirs <- ""
			return
		}

		n, err := dataConn.Read(f.dataBuf) 

		if err != nil {
			dirs <- "" 
			return 
		}

		dirs <- string(f.dataBuf[:n])
	}()

	for d := range dirs {
		close(dirs)
		return strings.Split(d, "\r\n"), nil
	}

	return nil, errors.New("failed to get directory listing")
}

func (f *FTP) Login(ftpLogin FTPLogin) (string, error) {

	userReq := "USER " + ftpLogin.username + "\r\n"

	_, err := f.conn.Write([]byte(userReq))

	if err != nil {
		return "", err
	}

	n, err := f.conn.Read(f.controlBuf)

	if err != nil {
		return "", nil
	}

	if string(f.controlBuf[:3]) == UserNameOkayNeedPassword {	

		passReq := "PASS " + ftpLogin.password + "\r\n"

		_, err := f.conn.Write([]byte(passReq)) 

		if err != nil {
			return "", err
		}

		n, err := f.conn.Read(f.controlBuf)

		if err != nil {
			return "", err
		}

		return string(f.controlBuf[:n-2]), nil
	} else {
		return string(f.controlBuf[:n-2]), nil
	}
}

func main() {	

	ftp := NewFTP(":")

	fmt.Println(ftp.Connect())

	fmt.Println(ftp.Login(FTPLogin{username: "ftpuser", password: "pass"}))

	dirs, err := ftp.List()

	if err != nil {
		log.Fatal(err)
	}

	for _, d := range dirs {
		fmt.Println(d)
	}
}