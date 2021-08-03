package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

type messageText struct {
	MsgID     int64  `json:"msgID"`
	Type      string `json:"type,omitempty"`
	TagSelf   string `json:"tag,omitempty"`
	TagClient string `json:"tagClient,omitempty"`
	Text      string `json:"text,omitempty"`
	Error     string `json:"error,omitempty"`
}

func makeResp(msg messageText) (b []byte, err error) {
	b, err = json.Marshal(msg)
	if err != nil {
		return
	}

	b, err = GenResponse(b)
	if err != nil {
		return
	}
	return
}

func GenResponse(b []byte) ([]byte, error) {
	b = []byte(hex.EncodeToString(b))
	var h, l uint8 = uint8(len(b) >> 8), uint8(len(b) & 0xff)
	r := []byte{byte(h), byte(l)}
	r = append(r, b...)
	return r, nil
}

func sendHello(c net.Conn, tagName string) error {
	var msg messageText

	msg.Type = "connect"
	msg.TagSelf = tagName

	return sendToServer(c, msg)
}

func sendError(c net.Conn, tagName string, errS string) error {
	var msg messageText

	msg.Type = "error"
	msg.TagSelf = tagName
	msg.Error = errS

	return sendToServer(c, msg)
}

func sendOk(c net.Conn, msg messageText) error {
	msg.Text = "ok"
	return sendToServer(c, msg)
}

func sendToServer(c net.Conn, msg messageText) error {
	r, err := makeResp(msg)
	if err != nil {
		return err
	}

	if n, err := c.Write(r); n == 0 || err != nil {
		return errors.New("error send error")
	}
	return nil
}

func startClient(tagName string, sendMsg bool, tagClient string) {
	dial := net.Dialer{Timeout: time.Second * 5}
	conn, err := dial.Dial("tcp", "localhost:65534")
	if err != nil {
		return
	}
	defer conn.Close()

	if err := sendHello(conn, tagName); err != nil {
		log.Printf("error connect: %v\n", err)
	}

	input := make([]byte, 1024)
	for {
		_, err := conn.Read(input)
		if err != nil {
			log.Printf("error read '%s': %v\n", tagName, err)
			return
		}

		lenData, err := strconv.ParseInt(hex.EncodeToString(input[:2]), 16, 64)
		if err != nil {
			sendError(conn, tagName, err.Error())
			continue
		}

		var ts string
		for i := 0; i < int(lenData); i++ {
			ts += string(input[i+2])
		}

		mbody, err := hex.DecodeString(ts)
		if err != nil {
			sendError(conn, tagName, err.Error())
			continue
		}

		var msg messageText
		if err := json.Unmarshal(mbody, &msg); err != nil {
			sendError(conn, tagName, err.Error())
			continue
		}

		conn.SetDeadline(time.Now().Add(180 * time.Second))

		if msg.Error != "" {
			log.Printf("error client '%s': %v\n", tagName, err)
			return
		}

		if msg.Type == "broadcast" {
			sendOk(conn, msg)
			continue
		}

		if msg.Type == "forward" {
			log.Printf("'%s' 'read forwarding msg': %v\n", tagName, msg)
			if msg.Text == "ok" {
				msg.TagClient = ""
				msg.TagSelf = tagName

			}
			sendOk(conn, msg)
			continue
		}

		if sendMsg {
			var newMsg messageText
			newMsg.TagSelf = tagName
			newMsg.TagClient = tagClient
			newMsg.Type = "forward"
			newMsg.Text = "send message"
			sendToServer(conn, newMsg)
		}

	}
}

func main() {
	for i := 2; i < 11; i++ {
		go startClient(fmt.Sprintf("client %d", i), false, "")
	}

	startClient("client 1", true, "client 2")
}
