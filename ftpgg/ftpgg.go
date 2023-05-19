package ftpgg

import (
	"fmt"
	"io"
	"net/textproto"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	Protocol    = "tcp"
	DefaultPort = "21"
)

const (
	ServiceReadyForNewUser       = 220
	UserNameOkayNeedPassword     = 331
	LoginSuccessful              = 230
	LoginIncorrect               = 530
	EnteringExtendedPassiveMode  = 229
	HereComesTheDirectoryListing = 150
	DirectorySendOk              = 226
	CurrentDirectoryOk           = 257
	SwitchingToBinaryMode        = 200
	FileSizeSent                 = 213
	OpeningFileInBinaryMode      = 150
	FileTransferComplete         = 226
	DirectoryChangedSuccessfully = 250
	OkToSendData                 = 150
	Goodbye                      = 221
	SiteChmodCommandOk           = 200
)

const (
	RegularFile EntryType = iota
	Directory
	Link
)

const (
	List       = "LIST"
	Pwd        = "PWD"
	Espv       = "EPSV"
	BinaryType = "TYPE I"
	Size       = "SIZE %s"
	Retr       = "RETR %s"
	Cwd        = "CWD %s"
	Stor       = "STOR %s"
	Quit       = "QUIT"
	Chmod      = "SITE chmod %s %s"
	Cdup       = "CDUP"
)

type EntryType int

type FTP struct {
	serverName string
	addr       string
	conn       *textproto.Conn
	dataConn   *textproto.Conn
}

type FTPLogin struct {
	Username string
	Password string
}

type Entry struct {
	Type        EntryType
	Name        string
	Date        time.Time
	Permissions string
}

func NewFTP(addr string) *FTP {
	return &FTP{
		addr: addr,
	}
}

func (f *FTP) Cmd(expected int, format string, args ...any) (int, string, error) {
	_, err := f.conn.Cmd(format, args...)
	if err != nil {
		return 0, "", err
	}
	return f.conn.ReadResponse(expected)
}

func (f *FTP) DataCmd(expected int, format string, args ...any) (int, string, error) {
	_, err := f.dataConn.Cmd(format, args...)
	if err != nil {
		return 0, "", err
	}
	return f.conn.ReadResponse(expected)
}

func (f *FTP) Connect() error {

	var err error

	f.conn, err = textproto.Dial(Protocol, f.addr+DefaultPort)

	if err != nil {
		return err
	}

	_, msg, err := f.conn.ReadResponse(ServiceReadyForNewUser)

	if err != nil {
		return err
	}

	f.serverName = msg

	return nil
}

func (f *FTP) enterPassiveMode() (string, error) {

	_, msg, err := f.Cmd(EnteringExtendedPassiveMode, Espv)

	if err != nil {
		return "", err
	}

	rgx := regexp.MustCompile("[0-9]+")

	return rgx.FindString(msg), nil
}

func (f *FTP) Download(fname string) ([]byte, error) {

	_, _, err := f.Cmd(SwitchingToBinaryMode, BinaryType)

	if err != nil {
		return nil, err
	}

	_, fsize, err := f.Cmd(FileSizeSent, Size, fname)

	if err != nil {
		return nil, err
	}

	nfsize, _ := strconv.Atoi(fsize)

	port, err := f.enterPassiveMode()

	if err != nil {
		return nil, err
	}

	f.dataConn, err = textproto.Dial(Protocol, f.addr+port)

	if err != nil {
		return nil, err
	}

	defer f.dataConn.Close()

	_, _, err = f.Cmd(OpeningFileInBinaryMode, Retr, fname)

	if err != nil {
		return nil, err
	}

	filebuf := make([]byte, nfsize)

	_, err = f.dataConn.R.Read(filebuf)

	if err != nil {
		return nil, err
	}

	_, _, err = f.conn.ReadResponse(FileTransferComplete)

	if err != nil {
		return nil, err
	}

	return filebuf, nil
}

func (f *FTP) Cdup() error {
	_, _, err := f.Cmd(DirectoryChangedSuccessfully, Cdup)
	return err
}

func (f *FTP) Cwd(path string) error {
	_, _, err := f.Cmd(DirectoryChangedSuccessfully, Cwd, path)
	return err
}

func (f *FTP) Stor(path string) error {

	port, err := f.enterPassiveMode()

	if err != nil {
		return err
	}

	f.dataConn, err = textproto.Dial("tcp", f.addr+port)

	if err != nil {
		return err
	}

	file, err := os.Open(path)

	if err != nil {
		return err
	}

	fstat, _ := file.Stat()

	defer file.Close()

	_, _, err = f.Cmd(OkToSendData, Stor, path)

	if err != nil {
		return err
	}

	uploaded := 0

	for {
		wr, _ := io.Copy(f.dataConn.W, file)
		uploaded += int(wr)
		if uploaded == int(fstat.Size()) {
			f.dataConn.Close()
			break
		}
	}

	_, _, err = f.conn.ReadResponse(FileTransferComplete)

	if err != nil {
		return err
	}

	return nil
}

func (f *FTP) Pwd() (string, error) {

	_, msg, err := f.Cmd(CurrentDirectoryOk, Pwd)

	if err != nil {
		return "", err
	}

	return msg, nil
}

func (f *FTP) Chmod(path string, mod string) error {
	_, _, err := f.Cmd(SiteChmodCommandOk, Chmod, mod, path)
	return err
}

func (f *FTP) Quit() error {
	_, _, err := f.Cmd(Goodbye, Quit)
	return err
}

func (f *FTP) List() ([]Entry, error) {

	port, err := f.enterPassiveMode()

	if err != nil {
		return nil, err
	}

	f.dataConn, err = textproto.Dial(Protocol, f.addr+port)

	if err != nil {
		return nil, err
	}

	defer f.dataConn.Close()

	_, _, err = f.Cmd(HereComesTheDirectoryListing, List)

	if err != nil {
		return nil, err
	}

	var entries []Entry

	for err != io.EOF {

		msg, err := f.dataConn.ReadLine()

		if err != nil {
			break
		}

		var entryType = RegularFile

		if msg[0] == 'd' {
			entryType = Directory
		} else if msg[0] == 'l' {
			entryType = Link
		}

		entry := strings.Fields(msg)
		date, _ := ParseDate(strings.Join([]string{entry[5], entry[6], entry[7]}, " "))

		entries = append(entries, Entry{
			Type:        entryType,
			Name:        entry[len(entry)-1],
			Permissions: entry[0],
			Date:        date,
		})
	}

	return entries, nil
}

func (f *FTP) Login(ftpLogin FTPLogin) error {

	_, _, err := f.Cmd(UserNameOkayNeedPassword, "USER %s", ftpLogin.Username)

	if err != nil {
		return err
	}

	_, _, err = f.Cmd(LoginSuccessful, "PASS %s", ftpLogin.Password)

	if err != nil {
		return err
	}

	return nil
}
