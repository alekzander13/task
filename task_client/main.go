package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {
	for i := 2; i < 11; i++ {
		go startClient(fmt.Sprintf("client %d", i), false, "")
	}

	startClient("client 1", true, "client 2")
}

func startClient(tagName string, sendMsg bool, tagClient string) {
	dial := net.Dialer{Timeout: time.Second * 5}
	conn, err := dial.Dial("tcp", "localhost:65534")
	if err != nil {
		return
	}
	defer conn.Close()

	str := "aa01" + hex.EncodeToString([]byte(tagName))

	b, err := GenResponse([]byte(str), false)
	if err != nil {
		return
	}

	log.Println(tagName, "write", b)
	if n, err := conn.Write(b); n == 0 || err != nil {
		return
	}

	input := make([]byte, 1024)
	for {
		lenR, err := conn.Read(input)
		if err != nil {
			log.Println(tagName, err)
			return
		}

		log.Println(tagName, "Read", input[:lenR])

		lenData, err := strconv.ParseInt(hex.EncodeToString(input[:2]), 16, 64)
		if err != nil {
			log.Println(tagName, err)
			return
		}

		var data []byte
		for i := 0; i < int(lenData); i++ {
			data = append(data, input[i+2])
		}

		if lenData > 3 {

		} else {
			var res string
			for i := 0; i < len(data); i++ {
				res += string(data[i])
			}

			tb, err := hex.DecodeString(res)
			if err != nil {
				return
			}

			if string(tb) == "bad" {
				return
			}
		}

		conn.SetDeadline(time.Now().Add(180 * time.Second))

		if sendMsg {
			bClientTag, err := GenResponse([]byte(tagClient), true)
			if err != nil {
				return
			}

			bTagName, err := GenResponse([]byte(tagName), true)
			if err != nil {
				return
			}

			msg := "aa02" + string(bClientTag) + string(bTagName) + hex.EncodeToString([]byte("sended msg"))

			b, err := GenResponse([]byte(msg), false)
			if err != nil {
				return
			}

			log.Println(tagName, "write", b)
			if n, err := conn.Write(b); n == 0 || err != nil {
				return
			}
		}
	}
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
