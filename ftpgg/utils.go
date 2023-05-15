package ftpgg

import (
	"bytes"
	"encoding/binary"
	"time"
)

func ParseDate(d string) (time.Time, error) {
	layout := "Jan 2 15:04"
	return time.Parse(layout, d)
}

func BytesToInt64(b []byte) (int64, error) {

	var n int64 

	if err := binary.Read(bytes.NewReader(b), binary.BigEndian, &n); err != nil {
		return -1, err
	}

	return n, nil
}