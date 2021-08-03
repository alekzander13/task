package main

import (
	"encoding/json"
	"net"
	"time"
)

type conn struct {
	net.Conn

	MsgID int64

	IdleTimeout   time.Duration
	MaxReadBuffer int64
	Tag           string
	Messages      map[int64]message
}

func (c *conn) Close() (err error) {
	err = c.Conn.Close()
	return
}

func (c *conn) updateDeadline() {
	idleDeadline := time.Now().Add(c.IdleTimeout)
	c.Conn.SetDeadline(idleDeadline)
}

func (c *conn) sendBadPacket() {
	r, err := GenResponse([]byte(`{"error":"bad request"}`))
	if err != nil {
		return
	}
	c.Conn.Write(r)
}

func (c *conn) sendPacket(msg messageText) error {
	c.MsgID++
	msg.MsgID = c.MsgID

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	r, err := GenResponse(b)
	if err != nil {
		return err
	}

	_, err = c.Conn.Write(r)
	if err == nil {
		if len(c.Messages) == 0 {
			c.Messages = make(map[int64]message)
		}
		c.Messages[msg.MsgID] = message{Text: r, Type: msg.Type, Count: 0}
		c.updateDeadline()
	}
	return err
}
