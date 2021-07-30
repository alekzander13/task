package main

import (
	"encoding/hex"
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

func (srv *Server) loop() {
	defer srv.mu.Unlock()
	for {
		srv.mu.Lock()
		for c := range srv.conns {
			c.sendGoodPacket()
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

func (srv *Server) sendDataToTagClient(tag string, data []byte) error {
	defer srv.mu.Unlock()
	srv.mu.Lock()
	var count int
	for c := range srv.conns {
		if c.Tag == tag {
			c.sendMessagePacket(data)
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
			fmt.Println(err)
			break
		}

		lenData, err := strconv.ParseInt(hex.EncodeToString(input[:2]), 16, 64)
		if err != nil {
			conn.sendBadPacket()
			continue
		}

		var data []byte
		for i := 0; i < int(lenData); i++ {
			data = append(data, input[i+2])
		}

		switch string(data[:4]) {
		case "aa01": //client intro
			for i := 4; i < len(data); i++ {
				if uint8(data[i]) > 0 {
					conn.Tag += string(data[i])
				}
			}

			tb, err := hex.DecodeString(conn.Tag)
			if err != nil {
				conn.sendBadPacket()
				break
			}
			conn.Tag = string(tb)
			if err := conn.sendGoodPacket(); err != nil {
				fmt.Println(err)
			}

		case "aa02": //Forwardin message
			var clientTag string
			tagLen, err := strconv.ParseInt(hex.EncodeToString(data[4:6]), 16, 64)
			if err != nil {
				conn.sendBadPacket()
				break
			}

			for i := 6; i < int(tagLen)+6; i++ {
				if uint8(data[i]) > 0 {
					clientTag += string(data[i])
				}
			}

			tb, err := hex.DecodeString(clientTag)
			if err != nil {
				fmt.Println(err)
				conn.sendBadPacket()
				break
			}

			clientTag = string(tb)
			srv.sendDataToTagClient(clientTag, data[tagLen+6:])

		default:
			conn.sendBadPacket()
		}
	}
}
