package ftpgg

import (
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
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
	Username string
	Password string
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

func (f *FTP) Download(fname string) ([]byte, error) {

	_, err := f.conn.Write([]byte("TYPE I\r\n"))

	if err != nil {
		return nil, err
	}	

	n, err := f.conn.Read(f.controlBuf)

	if err != nil {
		return nil, err
	}

	fmt.Println(string(f.controlBuf[:n]))

	_, err = f.conn.Write([]byte(fmt.Sprintf("SIZE %s\r\n", fname)))

	if err != nil {
		return nil, err
	}

	n, err = f.conn.Read(f.controlBuf)

	if err != nil {
		return nil, err
	}

	

	fsize, err := strconv.Atoi(string(f.controlBuf[4:n-2]))

	if err != nil {
		return nil, err
	}

	fmt.Printf("%d size in bytes", fsize)

	port, err := f.enterPassiveMode()

	if err != nil {
		return nil, err
	}

	dataConn, err := net.Dial("tcp", f.addr+port) 

	if err != nil {
		return nil, err
	}

	defer dataConn.Close()

	filebuf := make([]byte, fsize)

	_, err = f.conn.Write([]byte(fmt.Sprintf("RETR %s\r\n", fname)))

	if err != nil {
		return nil, err
	}

	n, err  = f.conn.Read(f.controlBuf)

	if err != nil {
		return nil, err
	}

	fmt.Println(string(f.controlBuf[:n]))	

	nb, err := dataConn.Read(filebuf) 
		
	if err != nil {
		return nil, err
	}

	fmt.Println(string(f.controlBuf[:n]))

	return filebuf[:nb], nil
}

func (f *FTP) Pwd() (string, error) {

	_, err := f.conn.Write(Pwd)

	if err != nil {
		return "", err
	}

	n, err := f.conn.Read(f.controlBuf)

	if err != nil {
		return "", err
	}

	return string(f.controlBuf[:n-2]), nil
}

func (f *FTP) List() ([]string, error) {

	port, err := f.enterPassiveMode()

	if err != nil  {
		return nil, err
	}

	dataConn, err := net.Dial("tcp", f.addr+port)

	if err != nil {
		return nil, err
	}

	defer dataConn.Close()

	_, err = f.conn.Write(List)

	if err != nil {
		return nil, err
	}

	_, err = io.CopyN(io.Discard, f.conn, 63)

	if err != nil {
		return nil, err
	}

	dirs := make(chan string, 1)

	defer close(dirs)

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
		return strings.Split(d, "\r\n"), nil
	}

	return nil, errors.New("failed to get directory listing")
}

func (f *FTP) Login(ftpLogin FTPLogin) (string, error) {

	userReq := "USER " + ftpLogin.Username + "\r\n"

	_, err := f.conn.Write([]byte(userReq))

	if err != nil {
		return "", err
	}

	n, err := f.conn.Read(f.controlBuf)

	if err != nil {
		return "", nil
	}

	if string(f.controlBuf[:3]) == UserNameOkayNeedPassword {	

		passReq := "PASS " + ftpLogin.Password + "\r\n"

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