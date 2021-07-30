package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//GetProgramPath used to create auxiliary files next to the executable file
func GetProgramPath() string {
	path := os.Args[0]
	p, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	path = filepath.Dir(p)
	ext := filepath.Ext(filepath.Base(p))
	p = strings.TrimSuffix(filepath.Base(p), ext)
	return filepath.Join(path, p)
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func ChkErrFatal(err error) {
	if err != nil {
		AddToLog(GetProgramPath()+"_errors.txt", err)
		os.Exit(1)
	}
}

func AddToLog(name string, info interface{}) {
	f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return
	}
	defer f.Close()

	now := time.Now()
	date := fmt.Sprintf("%02d.%02d.%02d %02d:%02d:%02d:%02d", now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond()) + " "

	fmt.Fprintln(f, date, info)
}

func GenResponse(b []byte, enc bool) ([]byte, error) {
	if enc {
		b = []byte(hex.EncodeToString(b))
	}
	var h, l uint8 = uint8(len(b) >> 8), uint8(len(b) & 0xff)
	r := []byte{byte(h), byte(l)}
	r = append(r, b...)
	return r, nil
}
