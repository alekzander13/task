package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	Addr         string
	IdleTimeout  time.Duration
	MaxReadBytes int64

	listener net.Listener
	conns    map[*conn]struct{}
	mu       sync.Mutex
}

func (srv *Server) ListenAndServe() {
	listen, err := net.Listen("tcp", ":"+srv.Addr)
	ChkErrFatal(err)

	log.Println("Server start on:", srv.Addr)

	defer listen.Close()

	srv.listener = listen

	go srv.loop()

	go srv.loopChkMsg()

	for {
		newConn, err := listen.Accept()
		if err != nil {
			AddToLog(GetProgramPath()+"_errors.txt", err)
			continue
		}

		c := &conn{
			Conn:          newConn,
			IdleTimeout:   srv.IdleTimeout,
			MaxReadBuffer: srv.MaxReadBytes,
		}

		srv.addConn(c)
		c.SetDeadline(time.Now().Add(c.IdleTimeout))
		go srv.handle(c)
	}
}

func (srv *Server) loopChkMsg() {
	defer srv.mu.Unlock()
	for {
		srv.mu.Lock()
		for c := range srv.conns {
			for k, v := range c.Messages {
				switch v.Type {
				case "connect":
					delete(c.Messages, k)
				case "broadcast":
					v.Count++
					if v.Count >= 3 {
						delete(c.Messages, k)
					} else {
						c.Write(v.Text)
						c.Messages[k] = v
					}
				case "forward":
					v.Count++
					if v.Count >= 3 {
						delete(c.Messages, k)
					} else {
						c.Write(v.Text)
						c.Messages[k] = v
					}
				}
			}
		}
		srv.mu.Unlock()
		time.Sleep(30 * time.Second)
	}
}

func (srv *Server) loop() {
	defer srv.mu.Unlock()
	for {
		srv.mu.Lock()
		for c := range srv.conns {
			var msg messageText
			msg.Type = "broadcast"
			msg.TagClient = c.Tag
			msg.Text = "ok"
			c.sendPacket(msg)
		}
		srv.mu.Unlock()
		time.Sleep(120 * time.Second)
	}
}

func (srv *Server) addConn(c *conn) {
	defer srv.mu.Unlock()
	srv.mu.Lock()
	if srv.conns == nil {
		srv.conns = make(map[*conn]struct{})
	}
	srv.conns[c] = struct{}{}
}

func (srv *Server) deleteConn(c *conn) {
	defer srv.mu.Unlock()
	srv.mu.Lock()
	delete(srv.conns, c)
}

func (srv *Server) sendDataToTagClient(msg messageText) error {
	defer srv.mu.Unlock()
	srv.mu.Lock()
	var count int
	for c := range srv.conns {
		if c.Tag == msg.TagClient {
			var newMsg messageText = msg
			newMsg.TagClient = msg.TagSelf
			newMsg.TagSelf = c.Tag
			if err := c.sendPacket(newMsg); err != nil {
				return err
			}
			count++
		}
	}
	if count > 0 {
		return nil
	}

	return errors.New("no clients")
}

func (srv *Server) handle(conn *conn) {
	defer func() {
		conn.Close()
		srv.deleteConn(conn)
	}()

	input := make([]byte, srv.MaxReadBytes)

	for {
		_, err := conn.Read(input)
		if err != nil {
			log.Println(err)
			break
		}

		lenData, err := strconv.ParseInt(hex.EncodeToString(input[:2]), 16, 64)
		if err != nil {
			conn.sendBadPacket()
			continue
		}

		var ts string
		for i := 0; i < int(lenData); i++ {
			ts += string(input[i+2])
		}

		mbody, err := hex.DecodeString(ts)
		if err != nil {
			conn.sendBadPacket()
			continue
		}

		var msg messageText
		if err := json.Unmarshal(mbody, &msg); err != nil {
			conn.sendBadPacket()
			continue
		}

		switch msg.Type {
		case "connect":
			conn.Tag = msg.TagSelf
			msg.Text = "ok"
			if err := conn.sendPacket(msg); err != nil {
				fmt.Println(err)
			}
		case "broadcast":
			delete(conn.Messages, msg.MsgID)
		case "forward":
			if msg.Text == "ok" {
				delete(conn.Messages, msg.MsgID)
			}
			if msg.TagClient != "" {
				srv.sendDataToTagClient(msg)
			}
		case "error":
			log.Printf("Error from %s text: %s", msg.TagSelf, msg.Error)

		default:
			conn.sendBadPacket()
		}
	}
}
